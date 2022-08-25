package framecountertime

var (
	Ticker  = make(chan int)
	UnixNow = int64(1651411507) // : This might be problematic if we load saves with offsets
)

func UpdateTicker(frame int) {
	if frame%60 == 0{
		go func() { Ticker <- frame }()
		UnixNow += 1
	}
}
