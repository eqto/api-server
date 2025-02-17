package api

type debugItem struct {
	msg string
}

type debugLog []debugItem

func (d *debugLog) log(msg string) {
	*d = append(*d, debugItem{msg})
}

func (d *debugLog) logErr(e error) {
	*d = append(*d, debugItem{e.Error()})
}

func (d debugLog) Strings() []string {
	strs := make([]string, len(d))
	for i, item := range d {
		strs[i] = item.msg
	}
	return strs
}
