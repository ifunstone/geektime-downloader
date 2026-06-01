package app

import "github.com/nicoxiang/geektime-downloader/internal/ui"

func toLegacyOption(option ProductTypeOption) ui.ProductTypeSelectOption {
	return ui.ProductTypeSelectOption{
		Index:              option.Index,
		Text:               option.Text,
		SourceType:         option.SourceType,
		AcceptProductTypes: option.AcceptProductTypes,
		NeedSelectArticle:  option.NeedSelectArticle,
		IsEnterpriseMode:   option.IsEnterpriseMode,
	}
}
