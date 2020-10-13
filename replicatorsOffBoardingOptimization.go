package main
import (
	"math/rand"
	"math"
	"fmt"
	"sort"
)

type Drive struct {
	Rating      float64
	Size		int
	Indices		[]int
}

type DynamicsAnswer struct {
	Rating      float64
	LastDrive	int
}

type OptimizationResult struct {
	Rating      float64
	Drives		[]int
}

type DrivesBatch struct {
	BigDrives		[]Drive
	NormalDrives	[]Drive
}

func RandBoundariesInt(minN int, maxN int) int {
	return rand.Intn(maxN - minN) + minN
}

func RandBoundariesFloat(minN float64, maxN float64) float64 {
	return rand.Float64()*(maxN - minN) + minN
}

func FindFunctionValueWithGivenParams(totalVolume int, needVolume int, maxRating float64) float64 {
	inf := 1e12
	if totalVolume < needVolume {
 		return float64(inf)
	}	else {
		var alpha float64 = 1
		return alpha * float64(totalVolume) / float64(needVolume) + maxRating
	}
}

func FindFunctionValue(drives []Drive, needVolume int) float64 {
	totalVolume := 0
	inf := 1e12
	maxRating := -inf
	for i := 0; i < len(drives); i++ {
		totalVolume += drives[i].Size
		maxRating = math.Max(maxRating, drives[i].Rating)
	}
	return FindFunctionValueWithGivenParams(totalVolume, needVolume, maxRating)
}

func GenerateRandomDrives(n int) []Drive {
	drives := make([]Drive, n)
	for i := 0; i < n; i++ {
		r := 10.0
		x := float64(RandBoundariesInt(0, 11))
		f := (r - x) / (r - 1.0)
		drives[i] = Drive{f * f *f, RandBoundariesInt(0, 100000), []int{i}}
	}
	return drives
}

func OptimizeWithGreed(drives []Drive, needVolume int) OptimizationResult {
	drivesBatch := DetachBigDrives(drives, needVolume)
	bestBigDrive := FindBestBigDrive(drivesBatch.BigDrives, needVolume)
	sort.SliceStable(drives, func(i, j int) bool {
    	return drives[i].Rating < drives[j].Rating
	})
	totalSize := 0
	selectedDrives := []int{}
	inf := 1e12
	maxPenalty := -inf
	for i := 0; totalSize < needVolume; i++ {
		selectedDrives = append(selectedDrives, i)
		totalSize += drives[i].Size
		maxPenalty = math.Max(maxPenalty, drives[selectedDrives[i]].Rating)
	}
	sort.SliceStable(selectedDrives, func(i, j int) bool {
    	return drives[selectedDrives[i]].Size > drives[selectedDrives[j]].Size
	})
	for i := len(selectedDrives) - 1; true; i-- {
		if totalSize - drives[selectedDrives[i]].Size >= needVolume {
			totalSize -= drives[selectedDrives[i]].Size
			selectedDrives = selectedDrives[:len(selectedDrives) - 1]
		} else {
			break
		}
	}
	selectedDrivesItems := []Drive{}
	for i := 0; i < len(selectedDrives); i++ {
		selectedDrivesItems = append(selectedDrivesItems, drives[selectedDrives[i]])
	}
	for i := 0; i < len(selectedDrives); i++ {
		selectedDrives[i] = drives[selectedDrives[i]].Indices[0]
	}
	funcValue := FindFunctionValue(selectedDrivesItems, needVolume)
	bestGreed := OptimizationResult{funcValue, selectedDrives}
	return UniteOptimizationResults(bestBigDrive, bestGreed)
}

func BruteForce(drives []Drive, level int, bestValue *float64, selectedDrives *[]int, alpha float64, curVolume int, needVolume int, maxDelta float64, curDrives []int) {
	optValue := alpha * float64(curVolume) / float64(needVolume) + maxDelta
	if optValue >= *bestValue {
		return
	}
	if curVolume >= needVolume {
		*bestValue = optValue
		*selectedDrives = curDrives
	} else if level < len(drives) {
		BruteForce(drives, level + 1, bestValue, selectedDrives, alpha, curVolume, needVolume, maxDelta, curDrives)
		curDrives = append(curDrives, level)
		BruteForce(drives, level + 1, bestValue, selectedDrives, alpha, curVolume + drives[level].Size, needVolume, math.Max(maxDelta, drives[level].Rating), curDrives)
		curDrives = curDrives[:len(curDrives) - 1]
	}
}

func (drive Drive) GetRoundedSize(scale int) int {
	return drive.Size / scale
}

func (group *Drive) AppendDrive(drive Drive) {
	group.Size += drive.Size
	group.Rating = math.Max(group.Rating, drive.Rating)
	group.Indices = append(group.Indices, drive.Indices[0])
}


func OptimizeWithBruteForce(drives []Drive, needVolume int) OptimizationResult {
	var alpha float64 = 1
	inf := 1e12
	bestValue := inf
	var selectedDrives []int
	var curDrives []int
	BruteForce(drives, 0, &bestValue, &selectedDrives, alpha, 0, needVolume, -inf, curDrives)
	return OptimizationResult{bestValue, selectedDrives}
}

func GetScale(volume int, drives int) int {
	operations := 100000000
	var scale int = volume * drives / operations
	if scale < 1 {
		scale = 1
	}
	return scale
}

func GenerateDrivesForDynamics(drives []Drive, scale int) []Drive {
	normalDrives := []Drive{}
	smallDrives := []Drive{}
	for i:= 0; i < len(drives); i++ {
		drives[i].GetRoundedSize(scale)
		scaled_size := drives[i].GetRoundedSize(scale)
		if scaled_size == 0 {
			smallDrives = append(smallDrives, drives[i])
		} else {
			normalDrives = append(normalDrives, drives[i])
		}
	}
	sort.SliceStable(smallDrives, func(i, j int) bool {
    	return smallDrives[i].Rating < smallDrives[j].Rating
	})
	inf := 1e12
	curDrive := Drive{-inf, 0, []int{}}
	for i := 0; i < len(smallDrives); i++ {
		curDrive.AppendDrive(smallDrives[i])
		if curDrive.GetRoundedSize(scale) > 0 {
			normalDrives = append(normalDrives, curDrive)
			curDrive = Drive{-inf, 0, []int{}}
		}
	}
	if curDrive.GetRoundedSize(scale) > 0 {
		normalDrives[len(normalDrives) - 1].AppendDrive(curDrive)
	}
	return normalDrives
}

func DetachBigDrives(drives []Drive, needVolume int) DrivesBatch {
	n := len(drives)
	bigDrives := []Drive{}
	normalDrives := []Drive{}
	for i := 0; i < n; i++ {
		if drives[i].Size >= needVolume {
			bigDrives = append(bigDrives, drives[i])
		}	else {
			normalDrives = append(normalDrives, drives[i])
		}
	}
	return DrivesBatch{bigDrives, normalDrives}
}

func FindBestBigDrive(drives []Drive, needVolume int) OptimizationResult {
	inf := 1e12
	bestArg := -1
	bestFuncValue := inf
	for i := 0; i < len(drives); i++ {
		value := FindFunctionValue([]Drive{drives[i]}, needVolume)
		if bestFuncValue > value {
			bestFuncValue = value
			bestArg = i
		}
	}
	if bestArg >= 0 {
		return OptimizationResult{bestFuncValue, drives[bestArg].Indices}
	}	else {
		return OptimizationResult{inf, []int{}}
	}
}

func UniteOptimizationResults(a OptimizationResult, b OptimizationResult) OptimizationResult {
	if a.Rating < b.Rating {
		return a
	}	else {
		return b
	}
}

func OptimizeWithDynamics(drives []Drive, needVolume int) OptimizationResult {
	drivesBatch := DetachBigDrives(drives, needVolume)
	bestBigDrive := FindBestBigDrive(drivesBatch.BigDrives, needVolume)
	scale := GetScale(needVolume, len(drives))
	groupedDrives := GenerateDrivesForDynamics(drivesBatch.NormalDrives, scale)
	scaledVolume := (needVolume + scale - 1) / scale 
	dp :=  make([]DynamicsAnswer, 2 * scaledVolume)
	inf := 1e12
	for i := 0; i < len(dp); i++ {
		dp[i] = DynamicsAnswer{inf, -1}
	}
	dp[0] = DynamicsAnswer{-inf, -1}
	n := len(groupedDrives)
	for i := 0; i < n; i++ {
		for j := scaledVolume - 1; j >= 0; j-- {
			if dp[j].Rating < inf {
				if dp[j + groupedDrives[i].GetRoundedSize(scale)].Rating > math.Max(dp[j].Rating, groupedDrives[i].Rating) {
					dp[j + groupedDrives[i].GetRoundedSize(scale)].Rating = math.Max(dp[j].Rating, groupedDrives[i].Rating)
					dp[j + groupedDrives[i].GetRoundedSize(scale)].LastDrive = i
				}
			}
		}
	}

	bestArg := -1
	bestValue := inf
	for i := scaledVolume; i < len(dp); i++ {
		if dp[i].Rating < inf {
			value := FindFunctionValueWithGivenParams(i, scaledVolume, dp[i].Rating)
			if value < bestValue {
				bestValue = value
				bestArg = i
			}
		}
	}
	var selectedDrives []int
	curArg := bestArg
	print(bestArg)
	for dp[curArg].LastDrive != -1 {
		selectedDrives = append(selectedDrives, groupedDrives[dp[curArg].LastDrive].Indices...)
		curArg -= groupedDrives[dp[curArg].LastDrive].GetRoundedSize(scale)
	}
	selectedDrivesItems := []Drive{}
	for i := 0; i < len(selectedDrives); i++ {
		selectedDrivesItems = append(selectedDrivesItems, drives[selectedDrives[i]])
	}
	funcValue := FindFunctionValue(selectedDrivesItems, needVolume)
	bestDynamics := OptimizationResult{funcValue, selectedDrives}
	return UniteOptimizationResults(bestBigDrive, bestDynamics)
}

func main(){
	rand.Seed(66282)
	drives := GenerateRandomDrives(100000)
	needVolume := 100000
	/*bruteForceBest := OptimizeWithBruteForce(drives, needVolume)
	fmt.Printf("%f\n", bruteForceBest.Rating)
	fmt.Printf("%v\n", bruteForceBest.Drives)*/
	dynamicsBest := OptimizeWithDynamics(drives, needVolume)
	fmt.Printf("Best Dynamics:\n")
	fmt.Printf("%f\n", dynamicsBest.Rating)
	fmt.Printf("%v\n", dynamicsBest.Drives)
	greedBest := OptimizeWithGreed(drives, needVolume)
	fmt.Printf("Best Greed:\n")
	fmt.Printf("%f\n", greedBest.Rating)
	fmt.Printf("%v\n", greedBest.Drives)
	best := UniteOptimizationResults(dynamicsBest, greedBest)
	fmt.Printf("Best:\n")
	fmt.Printf("%f\n", best.Rating)
	fmt.Printf("%v\n", best.Drives)
}