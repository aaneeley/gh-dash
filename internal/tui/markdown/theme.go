package markdown

import "charm.land/glamour/v2/ansi"

// DarkStyleConfig is the default dark style.
var CustomDarkStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "",
			BlockSuffix: "",
			Prefix:      "",
		},
		Margin:      uintPtr(0),
		IndentToken: stringPtr(""),
		Indent:      uintPtr(0),
	},
	Paragraph: ansi.StyleBlock{
		Margin:      uintPtr(0),
		IndentToken: stringPtr(""),
		Indent:      uintPtr(0),
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "",
			BlockSuffix: "",
			Prefix:      "",
		},
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{},
		Indent:         uintPtr(1),
		IndentToken:    stringPtr("│ "),
	},
	List: ansi.StyleList{
		LevelIndent: 2,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "",
			Color:       stringPtr("#89b4fa"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "# ",
			Color:  stringPtr("#cdd6f4"),
			Bold:   boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#cdd6f4"),
			Prefix: "## ",
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#cdd6f4"),
			Prefix: "### ",
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#bac2de"),
			Prefix: "#### ",
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#bac2de"),
			Prefix: "##### ",
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
			Color:  stringPtr("#a6adc8"),
			Bold:   boolPtr(false),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Bold: boolPtr(true),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#45475a"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: ansi.StylePrimitive{
		Color:       stringPtr("#89b4fa"),
		Underline:   boolPtr(false),
		BlockPrefix: "",
		BlockSuffix: "",
		Format:      "",
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#89b4fa"),
		Bold:  boolPtr(true),
	},
	Image: ansi.StylePrimitive{
		Underline: boolPtr(false),
		Color:     stringPtr("#89b4fa"),
		Bold:      boolPtr(false),
		Format:    "",
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#89b4fa"),
		Bold:   boolPtr(true),
		Format: "",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:       stringPtr("#a6e3a1"),
			Prefix:      "`",
			Suffix:      "`",
			BlockPrefix: "",
			BlockSuffix: "",
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#a6adc8"),
			},
			Margin: uintPtr(0),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#cdd6f4"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#11111b"),
				BackgroundColor: stringPtr("#f38ba8"),
			},
			Comment: ansi.StylePrimitive{
				Color: stringPtr("#6c7086"),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#fab387"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#cba6f7"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#cba6f7"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#f5c2e7"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#f9e2af"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#89dceb"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#9399b2"),
			},
			Name: ansi.StylePrimitive{
				Color: stringPtr("#cdd6f4"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#f38ba8"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#cba6f7"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#89b4fa"),
			},
			NameClass: ansi.StylePrimitive{
				Color:     stringPtr("#cdd6f4"),
				Underline: boolPtr(true),
				Bold:      boolPtr(true),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#fab387"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#89b4fa"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#fab387"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#a6e3a1"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#f5c2e7"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#f38ba8"),
			},
			GenericEmph: ansi.StylePrimitive{
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#a6e3a1"),
			},
			GenericStrong: ansi.StylePrimitive{
				Bold: boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#9399b2"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#11111b"),
			},
		},
	},
	Table: ansi.StyleTable{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Format:  " ",
				Conceal: boolPtr(true),
			},
		},
		CenterSeparator: stringPtr(""),
		ColumnSeparator: stringPtr(""),
		RowSeparator:    stringPtr(""),
	},
	DefinitionDescription: ansi.StylePrimitive{
		BlockPrefix: " ",
	},
}

func boolPtr(b bool) *bool { return &b }

func stringPtr(s string) *string { return &s }

func uintPtr(u uint) *uint { return &u }
