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
//const Query int = 6               // query size
//const Window int = 3              // window size
const activeUser int = 1 - 1      // activeuser index
const simThreshold float64 = 0.98 // 임계값

func main() {
	preferTS := make([][]info, 0)      // [i user][{movieid, timestamp}]
	preferItemTS := make([][][]int, 0) // [i user][set index][movies]
	candi := make([][]windows, 0)      // [i user][{sliding window index, disjoint window index}]
	sim := make([][]int, 0)

	genTS(&preferTS)                    // file로 부터 data 읽어 옴
	genItemTS(&preferTS, &preferItemTS) // timestamp 기준으로 집합을 나눔
	ssmCandi(&preferItemTS, &candi)     // 후보를 구함
	ssm(&preferItemTS, &candi, &sim)
}

func ssm(preferItemTS *[][][]int, candi *[][]windows, sim *[][]int) {
	simTmp := make([]int, 0)
	for i := 0; i < len(*preferItemTS); i++ {
		if i == activeUser {
			simTmp = simTmp[len(simTmp):]
			*sim = append(*sim, simTmp)
			continue
		}
		if len((*candi)[i]) == 0 {
			simTmp = simTmp[len(simTmp):]
			*sim = append(*sim, simTmp)
			continue
		}
		//for j:=0; j< len((*preferItemTS)[activeUser])

	}
}

func ssmCandi(preferItemTS *[][][]int, candi *[][]windows) { // window 간의 거리를 계산하여 후보 추출
	var Query int = len((*preferItemTS)[activeUser]) // Query size
	var Window int = (Query + 1) / 2                 // window size
	var p float64 = float64((Query+1)/Window - 1)    // p 구함
	//fmt.Println(Query, Window, p)
	sWindow := make([][]int, 0)
	for j := len((*preferItemTS)[activeUser]) - Window; j >= 0; j-- { // [idx][movieids...], movieid를 sliding window로 뒤에서 부터 구성
		sWindowTmp := make([]int, 0)
		for k := 0; k < Window; k++ {
			sWindowTmp = append(sWindowTmp, (*preferItemTS)[activeUser][j+k]...)
		}
		sWindow = append(sWindow, sWindowTmp)
	}
	candiTemp := make([]windows, 0)
	for i := 0; i < len(*preferItemTS); i++ {
		if i == activeUser {
			candiTemp = candiTemp[len(candiTemp):]
			*candi = append(*candi, candiTemp)
			continue
		}
		if len((*preferItemTS)[i]) < Query {
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
			for k := 0; k < len(sWindow); k++ { // k는 sliding window 갯수 만큼 반복
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
				if windowDistance <= (simThreshold / p) {
					candiTemp = append(candiTemp, windows{len((*preferItemTS)[activeUser]) - Window - k, len((*preferItemTS)[i]) - (j+1)*Window})
					//fmt.Println((*preferItemTS)[activeUser][len((*preferItemTS)[activeUser])-Window-k], (*preferItemTS)[i][len((*preferItemTS)[i])-(j+1)*Window])
					//fmt.Println(i+1, sWindow[k], dWindow[j])
					break
				}
			}
		}
		*candi = append(*candi, candiTemp)
	}

}

func genItemTS(preferTS *[][]info, preferItemTS *[][][]int) { // [i user][idx][movieIds...], timestamp를 기준으로 집합으로 나눈 i user의 idx 번째 집합의 element
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
