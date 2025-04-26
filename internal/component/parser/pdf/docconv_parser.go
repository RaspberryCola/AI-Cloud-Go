/*
基于 docconv 库的 pdf 解析器；
实现了Eino 组件接口的 Parse 方法。
*/

package pdf

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"

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
type DocconvPDFParser struct {
	ToPages bool
}

func NewDocconvPDFParser(ctx context.Context, config *Config) (*DocconvPDFParser, error) {
	if config == nil {
		config = &Config{}
	}
	return &DocconvPDFParser{ToPages: config.ToPages}, nil
}

func (pp *DocconvPDFParser) Parse(ctx context.Context, reader io.Reader, opts ...parser.Option) ([]*schema.Document, error) {
	// 1. 处理通用选项
	commonOpts := parser.GetCommonOptions(nil, opts...)

	specificOpts := parser.GetImplSpecificOptions(&options{
		toPages: &pp.ToPages,
	}, opts...)

	// 3. 实现解析逻辑
	fmt.Println("开始解析PDF文档...")
	res, meta, err := docconv.ConvertPDF(reader)
	if err != nil {
		fmt.Printf("PDF解析错误: %v\n", err)
		return nil, fmt.Errorf("PDF解析失败: %w", err)
	}

	fmt.Printf("PDF解析完成，文本长度: %d字符\n", len(res))
	fmt.Printf("PDF元数据: %+v\n", meta)

	// 检查解析结果是否为空
	if len(res) < 100 { // 至少需要100个字符才算有效
		fmt.Println("PDF解析结果太短或为空")
		if len(res) == 0 {
			return nil, fmt.Errorf("PDF解析结果为空，可能是扫描PDF或无文本内容")
		}
	}

	if *specificOpts.toPages {
		fmt.Println("待处理分页")
	}

	return []*schema.Document{{
		Content:  res,
		MetaData: commonOpts.ExtraMeta,
	}}, nil
}
