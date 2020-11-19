package btc

import (
	"errors"
	"math"
	"time"
)

// count in second
// use t.UnixNano() / int64(time.Millisecond) for milisecond
func makeTimestamp(t time.Time) int64 {
	return t.Unix()
}

// convert time.RFC3339 -> int64 value
// t,_ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
func makeTimestamp2(t string) (int64, error) {
	res, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return -1, err
	}
	return makeTimestamp(res), nil
}

// assume that each block will be produced in 10 mins ~= 600s
// this function will based on the given #param1 timestamp and #param3 chainTimestamp
// to calculate blockheight with approximate timestamp with #param1
// blockHeight = chainHeight - (chainTimestamp - timestamp) / 600
func estimateBlockHeight(self RandomClient, timestamp int64, chainHeight int, chainTimestamp int64, startTime time.Time, maxTime time.Duration) (int, error) {
	var estimateBlockHeight int
	// fmt.Printf("EstimateBlockHeight timestamp %d, chainHeight %d, chainTimestamp %d\n", timestamp, chainHeight, chainTimestamp)
	offsetSeconds := timestamp - chainTimestamp
	if offsetSeconds > 0 {
		return chainHeight, nil
	} else {
		estimateBlockHeight = chainHeight
		cacheDiff := 0
		isStart := false
		// diff is negative
		for true {
			if time.Since(startTime).Seconds() > maxTime.Seconds() {
				return -1, errors.New("estimate block height for random instruction exceed time out")
			}
			diff := int(offsetSeconds / (BTC_BLOCK_INTERVAL))
			if !isStart {
				cacheDiff = diff
				isStart = true
			} else {
				if math.Abs(float64(cacheDiff)) <= math.Abs(float64(diff)) {
					return estimateBlockHeight + cacheDiff, nil
				}
				cacheDiff = diff
			}
			estimateBlockHeight = estimateBlockHeight + diff
			//fmt.Printf("Estimate blockHeight %d \n", estimateBlockHeight)
			//if math.Abs(float64(diff)) < 5 {
			//	return estimateBlockHeight, nil
			//}
			blockTimestamp, _, err := self.GetTimeStampAndNonceByBlockHeight(estimateBlockHeight)
			// fmt.Printf("Try to estimate block with timestamp %d \n", blockTimestamp)
			if err != nil {
				return -1, err
			}
			if blockTimestamp == MaxTimeStamp {
				return -1, NewBTCAPIError(APIError, errors.New("Can't get result from API"))
			}
			offsetSeconds = timestamp - blockTimestamp
		}
	}
	return chainHeight, NewBTCAPIError(UnExpectedError, errors.New("Can't estimate block height"))
}
