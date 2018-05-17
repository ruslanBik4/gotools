// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import "fmt"

// проходим ряд чисел от 1 до 100
// вывести все факториалы чисел
func printFact() {

	fact := 1
	for i := 1; i < 100; i ++ {
		fact *= i
		fmt.Print(fact)
	}
}
// проходим ряд чисел от 1 до 100
// выводим на экран слово "Три" для тех,
// что целочисленно делятся на три
// выводим на экран слово "Пять" для тех,
// что целочисленно делятся на пять
func triPyat() {

	i3, i5 := 1, 1
	for i :=1; i < 100; i ++ {
		if i3 == 3 {
			fmt.Print("Три")
			i3 = 1
		}

		if i5 == 5 {
			fmt.Print("Три")
			i5 = 1
		}
		i3++
		i5++
	}
}
// Все программисты знают, что средний элемент в LinkedList несложно найти,
// определив длину списка, последовательно пройдя все его узлы,
// пока не дойдёшь до NULL в первом проходе. А затем, пройдя половину из них
// во втором проходе. Когда же их просят решить эту задачу за один проход,
// многие теряются.
type linkedList struct {
	data interface{}
	link *linkedList
}
func findAvgElement(list *linkedList) {
	link1, link2 := list, list

	i := 0
	for link1.link != nil {
		link1 = link1.link
		if i == 1 {
			link2 = link2.link
		}
		i = 1 - i
	}

}
//А если LinkedList зациклен?
func findAvgElementSave(list *linkedList) {
	link1, link2 := list, list

	i := 0
	for link1.link != nil {
		link1 = link1.link
		if i == 1 {
			link2 = link2.link
		}
		i = 1 - i
		if link1.link == link2.link {
			break
		}
	}

}

// Простейший триггер
// выдает значения 0, если передан 1
// 1, если передан 0
func trigger(a int) int {
	return 1 - a
}
// обмен значениями a и b (вариант для ГО)
func swap(a, b interface{})  {
	a, b = b, a
}