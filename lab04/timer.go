package main

	var n=3
	var ch=make([]chan int,n)

func timer(i int) {
	select {
	case x:=<-ch[i]:
		for j:=1;j<n;j++ {
			x<-ch[i]
		}
	default:
		for j:=0;j<n;j++ {
			ch[i]<-j
		}
	}
}

func main(){

	for i:=0;i<n;i++ {
		ch[i]=make(chan int)
	}
	for i:=0;i<n;i++ {
		go timer(i)
	}
}
