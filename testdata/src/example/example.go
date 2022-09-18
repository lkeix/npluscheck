package example

import "fmt"

type data struct{}

func (o *data) f5() {

}

func f() {
	var a interface{}
	var d data
	f4()
	f4()
	if true {
		f4()
	}
	for i := 0; i < 10; i++ {
		f1(a)
		d.f5()
		for i := 0; i < 100; i++ {
			s := f2()
			fmt.Println(s)
		}
		if true {
			f3()

			for i := 0; i < 10000; i++ {
				f4()
			}
		}
	}
}

func f1(c interface{}) {

}

func f2() string {
	return ""
}

func f3() {

}

func f4() {

}
