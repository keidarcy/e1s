package color

import "fmt"

var (
	GreenFmt = ""
	GrayFmt  = ""

	FooterSelectedItemFmt = ""
	FooterItemFmt         = ""
	FooterAwsFmt          = ""
	FooterE1sFmt          = ""

	HeaderTitleFmt = ""
	HeaderItemFmt  = ""
	HeaderKeyFmt   = ""

	HelpTitleFmt       = ""
	HelpKeyFmt         = ""
	HelpDescriptionFmt = ""

	NoticeInfoFmt  = ""
	NoticeWarnFmt  = ""
	NoticeErrorFmt = ""

	TableTitleFmt          = ""
	TableSecondaryTitleFmt = ""
	TableClusterTasksFmt   = ""
)

func (c Colors) initFmt() {
	GreenFmt = fmt.Sprintf("[%s]%%s[-:-:-]", c.Green)
	GrayFmt = fmt.Sprintf("[%s]%%s[-:-:-]", c.Gray)

	FooterSelectedItemFmt = fmt.Sprintf("[%s:%s:b] <%%s> [-:-:-]", c.Black, c.Cyan)
	FooterItemFmt = fmt.Sprintf("[%s:%s:] <%%s> [-:-:-]", c.Black, c.Gray)
	FooterAwsFmt = fmt.Sprintf("[%s:%s:bi] %%s ", c.Black, c.Yellow)
	FooterE1sFmt = fmt.Sprintf("[%s:%s:bi] %%s:%%s ", c.Black, c.Cyan)

	HeaderTitleFmt = fmt.Sprintf(" [%s]info([%s::b]%%s[%s:-:-]) ", c.Blue, c.Magenta, c.Blue)
	HeaderItemFmt = fmt.Sprintf(" %%s:[%s::b] %%s ", c.Cyan)
	HeaderKeyFmt = fmt.Sprintf(" [%s::b]<%%s> [%s:-:-]%%s ", c.Magenta, c.Green)

	HelpTitleFmt = fmt.Sprintf("[%s::b]%%s", c.Cyan)
	HelpKeyFmt = fmt.Sprintf("[%s::b]<%%s>", c.Magenta)
	HelpDescriptionFmt = fmt.Sprintf("[%s]%%s", c.Green)

	NoticeInfoFmt = fmt.Sprintf("✅ [%s::]%%s[-:-:-]", c.Green)
	NoticeWarnFmt = fmt.Sprintf("😔 [%s::]%%s[-:-:-]", c.Yellow)
	NoticeErrorFmt = fmt.Sprintf("💥 [%s::]%%s[-:-:-]", c.Red)

	TableTitleFmt = fmt.Sprintf(" [%s::-]<[%s::b]%%s[%s::-]>[%s::b]%%s[%s::-]([%s::b]%%d[%s::-]) ", c.Cyan, c.Magenta, c.Cyan, c.Cyan, c.Cyan, c.Magenta, c.Cyan)
	TableSecondaryTitleFmt = fmt.Sprintf(" [%s]%%s([%s::b]%%s[%s:-:-])[%s::-][[%s::-]%%s[-:-:-]] ", c.Blue, c.Magenta, c.Blue, c.FgColor, c.Green)
	TableClusterTasksFmt = fmt.Sprintf("[%s]%%d Pending[-] | [%s]%%d Running", c.Blue, c.Green)
}
