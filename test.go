package main

import (
	"fmt"
	"io"
	"os"
)

type info struct {
	movieId   int
	timestamp int
}

const dtime int = 1440

func main() {
	preferTS := make([][]info, 0)
	preferItemTS := make([][][]int, 0)
	genTS(&preferTS)
	genItemTS(&preferTS, &preferItemTS)
	fmt.Println(preferItemTS)
}

func genItemTS(preferTS *[][]info, preferItemTS *[][][]int) { // [i user][idx][movieIds...]
	movieIds := make([]int, 0)
	tmp := make([][]int, 0)
	for i := 0; i < len(*preferTS); i++ {
		movieIds = append(movieIds, (*preferTS)[i][0].movieId)
		for j := 0; j < len((*preferTS)[i])-1; j++ {
			if ((*preferTS)[i][j].timestamp / dtime) != ((*preferTS)[i][j+1].timestamp / dtime) {
				tmp = append(tmp, movieIds)
				movieIds = movieIds[len(movieIds):]
			}
			movieIds = append(movieIds, (*preferTS)[i][j+1].movieId)
		}
		tmp = append(tmp, movieIds)
		*preferItemTS = append(*preferItemTS, tmp)
		movieIds = movieIds[len(movieIds):]
		tmp = tmp[len(tmp):]
	}
}

func genTS(preferTS *[][]info) { // [i user][{movieId, timestamp}]
	var tmpInfo []info
	var userId, movieId, timestamp int
	var rating float64 // not used
	var preUserId int
	file, _ := os.Open("dat.txt")
	defer file.Close()
	fmt.Fscanln(file, &userId, &movieId, &rating, &timestamp)
	tmpInfo = append(tmpInfo, info{movieId, timestamp})
	preUserId = userId
	for {
		_, err := fmt.Fscanln(file, &userId, &movieId, &rating, &timestamp)
		if err == io.EOF {
			*preferTS = append(*preferTS, tmpInfo)
			break
		}
		if preUserId != userId {
			*preferTS = append(*preferTS, tmpInfo)
			tmpInfo = tmpInfo[len(tmpInfo):]
		}
		tmpInfo = append(tmpInfo, info{movieId, timestamp})
		preUserId = userId

	}
}
