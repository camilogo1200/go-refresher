package reverse

import (
	"fmt"
	"log"
	"strconv"
)

func Int(i int) int {
	itoaString := String(strconv.Itoa(i))
	i2, _ := strconv.Atoi(itoaString)
	log.Println(fmt.Sprintf("integer number = %d - itoa String = %s - reversed %d", i, itoaString, i2))
	return i2
}
