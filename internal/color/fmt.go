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

func (t Theme) initFmt() {
	GreenFmt = fmt.Sprintf("[%s]%%s[-:-:-]", t.Green)
	GrayFmt = fmt.Sprintf("[%s]%%s[-:-:-]", t.Gray)

	FooterSelectedItemFmt = fmt.Sprintf("[%s:%s:b] <%%s> [-:-:-]", t.Black, t.Cyan)
	FooterItemFmt = fmt.Sprintf("[%s:%s:] <%%s> [-:-:-]", t.Black, t.Gray)
	FooterAwsFmt = fmt.Sprintf("[%s:%s:bi] %%s ", t.Black, t.Yellow)
	FooterE1sFmt = fmt.Sprintf("[%s:%s:bi] %%s:%%s ", t.Black, t.Cyan)

	HeaderTitleFmt = fmt.Sprintf(" [%s]info([%s::b]%%s[%s:-:-]) ", t.Blue, t.Magenta, t.Blue)
	HeaderItemFmt = fmt.Sprintf(" %%s:[%s::b] %%s ", t.Cyan)
	HeaderKeyFmt = fmt.Sprintf(" [%s::b]<%%s> [%s:-:-]%%s ", t.Magenta, t.Green)

	HelpTitleFmt = fmt.Sprintf("[%s::b]%%s", t.Cyan)
	HelpKeyFmt = fmt.Sprintf("[%s::b]<%%s>", t.Magenta)
	HelpDescriptionFmt = fmt.Sprintf("[%s]%%s", t.Green)

	NoticeInfoFmt = fmt.Sprintf("✅ [%s::]%%s[-:-:-]", t.Green)
	NoticeWarnFmt = fmt.Sprintf("😔 [%s::]%%s[-:-:-]", t.Yellow)
	NoticeErrorFmt = fmt.Sprintf("💥 [%s::]%%s[-:-:-]", t.Red)

	TableTitleFmt = fmt.Sprintf(" [%s::-]<[%s::b]%%s[%s::-]>[%s::b]%%s[%s::-]([%s::b]%%d[%s::-]) ", t.Cyan, t.Magenta, t.Cyan, t.Cyan, t.Cyan, t.Magenta, t.Cyan)
	TableSecondaryTitleFmt = fmt.Sprintf(" [%s]%%s([%s::b]%%s[%s:-:-]) ", t.Blue, t.Magenta, t.Blue)
	TableClusterTasksFmt = fmt.Sprintf("[%s]%%d Pending[-] | [%s]%%d Running", t.Blue, t.Green)
}
