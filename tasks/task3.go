package tasks

import (
	"fmt"
	"runtime"
	"sync"
)

/*
3. Дана последовательность чисел: 2,4,6,8,10.
Найти сумму их квадратов(2*2+3*3+4*4….) с использованием конкурентных вычислений.

Реализация: пул воркеров с учетом Rate-лимитов (гибкая настройка ограничений ресурсов)
Синхронизация через wg.WaitGroup, wg.Lock/Unlock
*/

func Task3() {
	const (
		arrSize    = 5
		goroutines = 3
		quotaLimit = 2
	)

	array := [arrSize]int{2, 4, 6, 8, 10}
	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	workerInput := make(chan int, quotaLimit)
	var sum int

	// создаем горутины
	for i := 0; i < goroutines; i++ {
		go worker3(i, workerInput, wg, mutex, &sum)
	}

	// воркеры сами разберутся - задача идет в канал.
	// рандомная горутина принимает к работе
	for _, num := range array {
		// создаем счетчик задач для решения, чтобы дождаться корректного завершения горутин
		wg.Add(1)
		workerInput <- num
	}

	// обязательно закрыть канал (пул воркеров) - иначе не дождемся окончания
	// работы воркеров. Это может привести к дедлоку или утечки памяти
	close(workerInput)

	// ожидание завершения работы горутин
	wg.Wait()
	fmt.Printf("Task3. Sum of squares = %d\n\n", sum)
}

func worker3(workerName int, in <-chan int, wg *sync.WaitGroup, mutex *sync.Mutex, sum *int) {
	/*
		 используя пул воркеров, бесконечный цикл обязателен:
			-поскольку горутин создано 3, а задач в пуле воркеров - 5. Тогда
			3 горутины в сумме считают из канала 3 раза, а 2 задачи останутся необработанными
			Если завершить работу программы, оставив заполненным буфер канала (но не переполненным!),
			то паники (deadlock) - не произойдет (!),
			но задачи так и останутся не выполнеными. Вне зависимости от количества созданных
			горутин, в цикле горутина получит из канала новую задачу и выполнит ее

			но, используя бесконечный цикл, необходимо прервать его по завершению работы горутины
			чтобы не оставлять ее в памяти
	*/
	for {
		// num - значение из канала, а more - bool переменная, равная false, если канал закрыт
		// горутина завершается, если канал in закрыт
		num, more := <-in
		if more {
			// выполнение работы из пула
			mult := num * num

			// Чтобы избежать состояния гонки - блокируем горутины
			mutex.Lock()
			*sum += mult

			// т.к. выводим значение с общим ресурсом *sum, то Println() - внутри Lock()
			fmt.Printf("Goroutine #%d: sqr(%d) = %d, sum = %d \n", workerName, num, mult, *sum)
			mutex.Unlock()

			/*
				уменьшаем счетчик задач для решения.
				defer - использовать обязательно (!), т.к. если оставить wg.Done() без defer,
				то после выполнения всех задач сразу выполнится wg.Wait(), который не станет ждать завершения горутин,
				т.к. счетчик выполненных задач станет равным нулю. Будет выведен результат. А сообщение о завершении
				горутин - нет.
				т.е. горутины не успеют завершиться до окончания работы программы. Завершиться - выполнить ветку else с
				явным возвратом return (и выводом на экран уведомления о завершении работы)
			*/
			defer wg.Done()
			runtime.Gosched()
		} else {
			fmt.Printf("Task3. All jobs is done. (Worker #%d) \n", workerName)
			return
		}
	}
}