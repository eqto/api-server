package api

type debugItem struct {
	msg string
}

type debugLog []debugItem

func (d debugLog) log(msg string) {
	d = append(d, debugItem{msg})
}

func (d debugLog) logErr(e error) {
	d = append(d, debugItem{e.Error()})
}
