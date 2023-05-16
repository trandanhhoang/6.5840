package main

import (
	"fmt"
	"sync"
)

func main(){
	litString := []string{"1","2","3","4","5","6","7"}

	var done sync.WaitGroup

	for _,s := range litString{
		done.Add(1)
		go func(s string){
			/* may be fmt.Println error and we can't reach the done.Done()
				fmt.Println(s)
				done.Done()
			*/
			// -> must using defer
			defer done.Done()
			fmt.Println(s)
		}(s)
	}
	done.Wait()
}