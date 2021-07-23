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

type windows struct {
	slidingIndex  int
	disjointIndex int
}

const dtime int = 1440 // 집합 나누는 기준, timestamp 기준
//const activeUser int = 1 - 1      // activeuser index
const minQuery int = 4
const setSize int = 20

func main() {
	var choose int
	var activeUser, query, queryIndex int
	var threshold float64
	preferTS := make([][]info, 0)      // [i user][{movieid, timestamp}]
	preferItemTS := make([][][]int, 0) // [i user][set index][movies]
	candi := make([][]windows, 0)      // [i user][{sliding window index, disjoint window index}]
	sim := make([][]int, 0)

	genTS(&preferTS) // file로 부터 data 읽어 옴
	fmt.Print("집합 구성 방식 1. timestamp 2. setSize : ")
	fmt.Scanln(&choose)
	if choose == 1 {
		genItemTS_timestamp(&preferTS, &preferItemTS) // timestamp 기준으로 집합을 나눔
	}
	if choose == 2 {
		genItemTS_setSize(&preferTS, &preferItemTS) // setSize 기준으로 집합을 나눔
	}
	//fmt.Println(len(preferTS[1]))
	fmt.Print("activeUser ( 1 ~ ", len(preferTS), " ) : ")
	fmt.Scanln(&activeUser)
	activeUser--
	fmt.Print("query size ( ", minQuery, " ~ ", len(preferItemTS[activeUser]), " ) : ")
	fmt.Scanln(&query)
	fmt.Print("query index ( 0 ~ ", len(preferItemTS[activeUser])-query, " ) : ")
	fmt.Scanln(&queryIndex)
	fmt.Print("threshold : ")
	fmt.Scanln(&threshold)
	ssmCandi(activeUser, query, queryIndex, threshold, &preferItemTS, &candi) // 후보를 구함
	ssm(activeUser, query, threshold, &preferTS, &candi, &sim)                // 전체 매칭
	/*
		for i := 0; i < len(preferItemTS); i++ {
			for j := 0; j < len(preferItemTS[i]); j++ {
				fmt.Println(j+1, preferItemTS[i][j])
			}
			fmt.Println("")
		}

		for i := 0; i < len(preferTS[0]); i++ {
			for j := 0; j < len(preferTS[1]); j++ {
				if preferTS[0][i].movieId == preferTS[1][j].movieId {
					fmt.Print(preferTS[0][i].movieId, " ")
				}
			}
		}
	*/
}

func ssm(activeUser int, query int, threshold float64, preferTS *[][]info, candi *[][]windows, sim *[][]int) { // refine 미완성
	for i := 0; i < len((*candi)); i++ {
		if len((*candi)[i]) != 0 {
			fmt.Println(i + 1)
		}
	}

	/*
		active := make([]int, 0)
		for i := 0; i < len((*preferTS)[activeUser]); i++ {
			active = append(active, (*preferTS)[activeUser][i].movieId)
		}
		candiUser := make([][]int, 0)
		for i := 0; i < len(*preferTS); i++ {
			candiUserTmp := make([]int, 0)
			if i == activeUser {
				candiUserTmp = candiUserTmp[len(candiUserTmp):]
				candiUser = append(candiUser, candiUserTmp)
				continue
			}
			if len((*candi)[i]) == 0 {
				candiUserTmp = candiUserTmp[len(candiUserTmp):]
				candiUser = append(candiUser, candiUserTmp)
				continue
			}
			for j := 0; j < len((*preferTS)[i]); j++ {
				candiUserTmp = append(candiUserTmp, (*preferTS)[i][j].movieId)
			}
			candiUser = append(candiUser, candiUserTmp)
			candiUserTmp = candiUserTmp[len(candiUserTmp):]
		}
		for i := 0; i < len(candiUser); i++ {
			var distance float64
			var setJac, intersec float64
			for j := 0; j < len(active); j++ {
				for k := 0; k < len(candiUser[i]); k++ {
					if active[j] == candiUser[i][k] {
						intersec++
					}
				}
			}
			setJac = intersec / (float64(len(active)) + float64(len(candiUser[i])) - intersec)
			distance += (1 - setJac)
			if distance <= threshold {
				fmt.Println(i + 1)
			}
		}
	*/
}

func ssmCandi(activeUser int, query int, queryIndex int, threshold float64, preferItemTS *[][][]int, candi *[][]windows) { // window 간의 거리를 계산하여 후보 추출
	var Window int = (minQuery + 1) / 2                 // 최대 window size 구함, (min(Q) + 1) / 2
	var p float64 = float64(((query + 1) / Window) - 1) // p 구함, (len(Q)+1/w) - 1
	fmt.Println("Window", Window, "p", p, "threshold/p", threshold/p)

	// sliding window 구성
	sWindow := make([][]int, 0)
	for j := queryIndex + query - Window; j >= queryIndex; j-- {
		sWindowTmp := make([]int, 0)
		for k := 0; k < Window; k++ {
			sWindowTmp = append(sWindowTmp, (*preferItemTS)[activeUser][j+k]...)
		}
		sWindow = append(sWindow, sWindowTmp)
	}

	// disjoint window 구성
	candiTemp := make([]windows, 0)
	for i := 0; i < len(*preferItemTS); i++ {
		/*	// activeUser를 제거 하고 검사 할 수 있다
			if i == activeUser {
				candiTemp = candiTemp[len(candiTemp):]
				*candi = append(*candi, candiTemp)
				continue

			}
		*/
		if len((*preferItemTS)[i]) < minQuery {
			candiTemp = candiTemp[len(candiTemp):]
			*candi = append(*candi, candiTemp)
			continue
		}
		candiTemp = candiTemp[len(candiTemp):]
		dWindow := make([][]int, 0)
		for j := len((*preferItemTS)[i]) - Window; j >= 0; j -= Window { // [idx][movieids...], movieid를 disjoint window로 뒤에서 부터 구성
			dWindowTmp := make([]int, 0)
			for k := 0; k < Window; k++ {
				dWindowTmp = append(dWindowTmp, (*preferItemTS)[i][j+k]...)
			}
			dWindow = append(dWindow, dWindowTmp)
		}
		for j := 0; j < len(dWindow); j++ { // j는 disjoint window 갯수 만큼 반복
			for k := 0; k < len(sWindow); k++ { // k는 sliding window 갯수 만큼

				var windowDistance float64
				var setJac, intersec float64
				for l := 0; l < len(dWindow[j]); l++ { // l은 disjoint window의 movieid의 수만큼 반복
					for m := 0; m < len(sWindow[k]); m++ { // m은 sliding window의 movieid의 수만큼 반복
						if dWindow[j][l] == sWindow[k][m] {
							intersec++ // 교집합 갯수 증가
						}
					}
				}
				setJac = intersec / (float64(len(dWindow[j])) + float64(len(sWindow[k])) - intersec) // 교집합 갯수로 jaccard coefficient
				windowDistance += (1 - setJac)                                                       // window간 거리를 구함
				/*
					if i == 0 {
						fmt.Print(" ", windowDistance)
					}
				*/
				if windowDistance <= (threshold / p) {
					candiTemp = append(candiTemp, windows{len((*preferItemTS)[activeUser]) - Window - k, len((*preferItemTS)[i]) - (j+1)*Window})
					/*
						fmt.Print(i+1, (*preferItemTS)[activeUser][len((*preferItemTS)[activeUser])-Window-k], ">")
						for l := 0; l < Window; l++ {
							fmt.Print((*preferItemTS)[i][len((*preferItemTS)[i])-(j+1)*Window+l])
						}
						fmt.Println("")
					*/
					//fmt.Println(i+1, sWindow[k], dWindow[j])
					//break
				}
			}
			//fmt.Println("")
		}
		*candi = append(*candi, candiTemp)
	}

}

// setSize 크기로 집합을 구성한다.
// 집합의 크기로 나눴을 때 나머지를 제거 하는데, 최신 element가 제거 되는 것을 막기 위해 앞부분의 나머지를 제거한다
func genItemTS_setSize(preferTS *[][]info, preferItemTS *[][][]int) { // [i user][idx][movieIds...], timestamp를 기준으로 집합으로 나눈 i user의 idx 번째 집합의 element
	movieIds := make([]int, 0)
	tmp := make([][]int, 0)
	for i := 0; i < len(*preferTS); i++ {
		var count int
		idx := len((*preferTS)[i]) % setSize
		for j := idx; j < len((*preferTS)[i]); j++ {
			movieIds = append(movieIds, (*preferTS)[i][j].movieId)
			count++
			if count == setSize {
				tmp = append(tmp, movieIds)
				movieIds = movieIds[len(movieIds):]
				count = 0
			}
		}
		*preferItemTS = append(*preferItemTS, tmp)
		tmp = tmp[len(tmp):]
	}
	// 집합 구성 check
	/*
		for i := 0; i < len((*preferItemTS)); i++ {
			fmt.Println(len((*preferTS)[i]))
			for j := 0; j < len((*preferItemTS)[i]); j++ {
				fmt.Println((*preferItemTS)[i][j], len((*preferItemTS)[i][j]))
			}
			fmt.Println(len((*preferItemTS)[i]))
			fmt.Println("")
		}
	*/
}

// timestamp를 기준으로 집합을 구성한다
// 집합마다 size가 다르다
func genItemTS_timestamp(preferTS *[][]info, preferItemTS *[][][]int) { // [i user][idx][movieIds...], timestamp를 기준으로 집합으로 나눈 i user의 idx 번째 집합의 element
	movieIds := make([]int, 0)
	tmp := make([][]int, 0)
	for i := 0; i < len(*preferTS); i++ {
		movieIds = append(movieIds, (*preferTS)[i][0].movieId)
		for j := 0; j < len((*preferTS)[i])-1; j++ {
			if ((*preferTS)[i][j].timestamp / dtime) != ((*preferTS)[i][j+1].timestamp / dtime) { // 이전 집합과의 비교
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

// 파일로 부터 data를 읽어온다
func genTS(preferTS *[][]info) { // [i user][{movieId, timestamp}], file로 부터 userid, movieid, timestamp 읽어 옴
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
