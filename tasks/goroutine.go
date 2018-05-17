// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"sync"
)
// вопрос - что выведет программа?
var list = []int{1,3,5,6,10}
func main() {
	fmt.Println( "start")

	wGroup := &sync.WaitGroup{}

	for i, val := range list {
		go func(task int) {
			wGroup.Add(1)
			defer wGroup.Done()
			fmt.Println(i*val)
		}(i)
	}

	wGroup.Wait()
	fmt.Print( "finished")
}
// еще вариант
// здесь убраны ошибки предыдущей
func main1() {
	fmt.Println( "start")

	wGroup := &sync.WaitGroup{}
	broadcast := make(chan struct{})

	for i, val := range list {
		wGroup.Add(1)
		go func(i, val int) {
			defer wGroup.Done()
			fmt.Println(i*val)
			for fl, ok := <- broadcast; ok; {
				fmt.Println(fl)
			}

		}(i, val)
	}

	broadcast <- nil

	wGroup.Wait()
	fmt.Print( "finished")
	close(broadcast)
}
// сюда же оп. вопрос - как сделать, чтобы все гоурутины сделали принт?