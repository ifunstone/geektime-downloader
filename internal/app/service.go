package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/course"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/progress"
)

type ProductTypeOption struct {
	Index              int
	Text               string
	SourceType         int
	AcceptProductTypes []string
	NeedSelectArticle  bool
	IsEnterpriseMode   bool
}

type DirectVideoProduct struct {
	Title      string
	ArticleID  int
	SourceType int
}

type SelectionResult struct {
	ProductType  ProductTypeOption
	Course       geektime.Course
	DirectVideo  *DirectVideoProduct
	IsDirectMode bool
}

type Service struct {
	ctx    context.Context
	cfg    *config.AppConfig
	client *geektime.Client
	progress progress.Func
}

func DefaultConfig() config.AppConfig {
	userHomeDir, _ := os.UserHomeDir()
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)

	return config.AppConfig{
		DownloadFolder:         defaultDownloadFolder,
		Quality:                "sd",
		DownloadComments:       1,
		ColumnOutputType:       1,
		PrintPDFWaitSeconds:    5,
		PrintPDFTimeoutSeconds: 60,
		Interval:               1,
		LogLevel:               "info",
	}
}

func ProductTypeOptions(isEnterprise bool) []ProductTypeOption {
	if isEnterprise {
		return []ProductTypeOption{
			{Index: 0, Text: "训练营", SourceType: 5, AcceptProductTypes: []string{"c44"}, NeedSelectArticle: true, IsEnterpriseMode: true},
		}
	}

	return []ProductTypeOption{
		{Index: 0, Text: "普通课程", SourceType: 1, AcceptProductTypes: []string{"c1", "c3"}, NeedSelectArticle: true},
		{Index: 1, Text: "每日一课", SourceType: 2, AcceptProductTypes: []string{"d"}, NeedSelectArticle: false},
		{Index: 2, Text: "公开课", SourceType: 1, AcceptProductTypes: []string{"p35", "p29", "p30"}, NeedSelectArticle: true},
		{Index: 3, Text: "大厂案例", SourceType: 4, AcceptProductTypes: []string{"q"}, NeedSelectArticle: false},
		{Index: 4, Text: "训练营", SourceType: 5, AcceptProductTypes: []string{""}, NeedSelectArticle: true},
		{Index: 5, Text: "其他", SourceType: 1, AcceptProductTypes: []string{"x", "c6"}, NeedSelectArticle: true},
	}
}

func (p ProductTypeOption) IsUniversity() bool {
	return p.Index == 4 && !p.IsEnterpriseMode
}

func NewService(ctx context.Context, cfg *config.AppConfig, progressFn progress.Func) (*Service, error) {
	if err := config.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	logger.Init(cfg.LogLevel)
	client := geektime.NewClient(config.ReadCookiesFromInput(cfg))
	return &Service{
		ctx:    ctx,
		cfg:    cfg,
		client: client,
		progress: progressFn,
	}, nil
}

func (s *Service) ResolveProduct(productType ProductTypeOption, productID int) (*SelectionResult, error) {
	if productType.NeedSelectArticle {
		courseInfo, err := s.loadCourse(productType, productID)
		if err != nil {
			return nil, err
		}
		return &SelectionResult{
			ProductType: productType,
			Course:      courseInfo,
		}, nil
	}

	productInfo, err := s.client.ProductInfo(productID)
	if err != nil {
		return nil, err
	}

	if productInfo.Data.Info.Extra.Sub.AccessMask == 0 {
		return nil, fmt.Errorf("尚未购买该课程")
	}

	if !s.validateProductCode(productType, productInfo.Data.Info.Type) {
		return nil, fmt.Errorf("输入的课程 ID 有误")
	}

	return &SelectionResult{
		ProductType: productType,
		Course: geektime.Course{
			Access:  true,
			ID:      productID,
			Title:   productInfo.Data.Info.Title,
			Type:    productInfo.Data.Info.Type,
			IsVideo: true,
		},
		DirectVideo: &DirectVideoProduct{
			Title:      productInfo.Data.Info.Title,
			ArticleID:  productInfo.Data.Info.Article.ID,
			SourceType: productType.SourceType,
		},
		IsDirectMode: true,
	}, nil
}

func (s *Service) DownloadAll(result *SelectionResult) error {
	downloader := course.NewCourseDownloader(s.ctx, s.cfg, s.client, nil, s.progress)

	if result.IsDirectMode && result.DirectVideo != nil {
		if s.progress != nil {
			s.progress(progress.Download{
				CourseTitle: result.Course.Title,
				ItemTitle:   result.DirectVideo.Title,
				Stage:       "starting",
				CurrentItem: 1,
				TotalItems:  1,
			})
		}
		return downloader.DownloadSingleVideoProduct(
			result.DirectVideo.Title,
			result.DirectVideo.ArticleID,
			result.DirectVideo.SourceType,
		)
	}

	return downloader.DownloadAll(result.Course, toLegacyOption(result.ProductType))
}

func (s *Service) DownloadArticle(result *SelectionResult, articleIndex int) error {
	if articleIndex < 0 || articleIndex >= len(result.Course.Articles) {
		return fmt.Errorf("请选择有效的文章")
	}

	if result.IsDirectMode {
		return fmt.Errorf("当前课程类型不支持选择单篇")
	}

	article := result.Course.Articles[articleIndex]
	if result.ProductType.IsUniversity() {
		universityArticleDetail, err := s.client.UniversityClassArticleDetail(result.Course.ID, article.AID)
		if err != nil {
			return err
		}
		if universityArticleDetail.Data.VideoID == "" {
			return fmt.Errorf("训练营暂时只支持下载视频")
		}
	}

	downloader := course.NewCourseDownloader(s.ctx, s.cfg, s.client, nil, s.progress)
	return downloader.DownloadArticle(result.Course, toLegacyOption(result.ProductType), article, true)
}

func (s *Service) loadCourse(productType ProductTypeOption, productID int) (geektime.Course, error) {
	if productType.IsEnterpriseMode {
		courseInfo, err := s.client.EnterpriseCourseInfo(productID)
		if err != nil {
			return geektime.Course{}, err
		}
		if !courseInfo.Access {
			return geektime.Course{}, fmt.Errorf("尚未购买该课程")
		}
		return courseInfo, nil
	}

	if productType.IsUniversity() {
		courseInfo, err := s.client.UniversityClassInfo(productID)
		if err != nil {
			return geektime.Course{}, err
		}
		if !courseInfo.Access {
			return geektime.Course{}, fmt.Errorf("尚未购买该课程")
		}
		return courseInfo, nil
	}

	courseInfo, err := s.client.CourseInfo(productID)
	if err != nil {
		return geektime.Course{}, err
	}
	if !courseInfo.Access {
		return geektime.Course{}, fmt.Errorf("尚未购买该课程")
	}
	if !s.validateProductCode(productType, courseInfo.Type) {
		return geektime.Course{}, fmt.Errorf("输入的课程 ID 有误")
	}
	return courseInfo, nil
}

func (s *Service) validateProductCode(productType ProductTypeOption, productCode string) bool {
	for _, acceptedType := range productType.AcceptProductTypes {
		if acceptedType == productCode {
			return true
		}
	}
	return false
}
