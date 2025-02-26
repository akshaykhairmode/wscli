package global

var stopApp = make(chan struct{}, 2)

func Stop() {
	stopApp <- struct{}{}
}

func WaitForStop() {
	<-stopApp
}
