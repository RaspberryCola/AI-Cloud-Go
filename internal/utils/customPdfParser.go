package utils

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	"io"

	"code.sajari.com/docconv/v2"
)

// options
// 定制实现自主定义的 option 结构体
type options struct {
	toPages *bool
}

func WithToPages(toPages bool) parser.Option {
	return parser.WrapImplSpecificOptFn(func(opts *options) {
		opts.toPages = &toPages
	})
}

type Config struct {
	ToPages bool
}
type CustomPdfParser struct {
	ToPages bool
}

func NewCustomPdfParser(ctx context.Context, config *Config) (*CustomPdfParser, error) {
	if config == nil {
		config = &Config{}
	}
	return &CustomPdfParser{ToPages: config.ToPages}, nil
}

func (pp *CustomPdfParser) Parse(ctx context.Context, reader io.Reader, opts ...parser.Option) ([]*schema.Document, error) {
	// 1. 处理通用选项
	commonOpts := parser.GetCommonOptions(nil, opts...)

	specificOpts := parser.GetImplSpecificOptions(&options{
		toPages: &pp.ToPages,
	}, opts...)

	// 3. 实现解析逻辑
	res, _, _ := docconv.ConvertPDF(reader)

	if *specificOpts.toPages {
		fmt.Println("待处理分页")
	}

	return []*schema.Document{{
		Content:  res,
		MetaData: commonOpts.ExtraMeta,
	}}, nil
}
