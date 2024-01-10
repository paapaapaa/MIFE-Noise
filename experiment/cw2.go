package main

import (
	"bufio"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fentec-project/gofe/data"
	"github.com/fentec-project/gofe/innerprod/simple"
)

// The structure holding a single node/leaf of the binary tree, Range is the range of values, Value is occurences of the Range in the dataset
type Node struct {
	Range [2]float32
	Value int
}

// basic function to read the dataset files.
func readFile(filename string) ([]float32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var floats []float32
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		value, err := strconv.ParseFloat(line, 32)
		if err != nil {
			return nil, err
		}
		floats = append(floats, float32(value))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return floats, nil
}

// Go seemingly doesn't have native min() or max() functions for slices, so this one provides both
func findMinMax(dataset []float32) (float32, float32) {

	min := dataset[0]
	max := dataset[0]

	for _, value := range dataset {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	return min, max

}

// sampling laplace noise for each dataset entry following the formula Y = µ - b sgn(u)ln(1-2|u|)
func AddLaplaceNoise(dataset []*Node, mu, b float64) {

	for i := range dataset {
		u := rand.Float64() - 0.5 // Uniform random value in [-0.5, 0.5]
		sign := float64(1)
		if u < 0 {
			sign = -1
		}

		noise := mu - b*sign*math.Log(1-2*math.Abs(u))
		dataset[i].Value += int(noise) // the noise needs to be converted to Int, because the gofe library cannot take float values.
	}
}

// this one creates a random query of the form "how many values lie between [min,max]",
// then it returns the values by two different tree traversal functions
func query(tree []*Node) ([]int, []int) {

	maxTotal := tree[0].Range[1]

	terribleIndices := []int{}

	min := rand.Float32() * maxTotal
	max := rand.Float32() * maxTotal

	if min > max {
		cache := min
		min = max
		max = cache
	}

	terribleIndices = append(terribleIndices, traverseBottom(tree, 0, [2]float32{min, max}, len(tree))...) //only leaves
	niceIndices := traverseNearest(tree, 0, [2]float32{min, max}, len(tree))                               // lowest single range that encapsulates [min,max]

	return terribleIndices, []int{niceIndices}
}

// recursive function that finds the indices of the leaves that encapsulate given range [a,b],
// only returns leaves, but most accurate when query is between ranges in the tree
func traverseBottom(tree []*Node, i int, Range [2]float32, maxNodes int) []int {

	if 2*i+1 >= maxNodes {
		return []int{i}
	} else {

		vals := []int{}

		if Range[0] < tree[2*i+1].Range[1] {
			vals = append(vals, traverseBottom(tree, 2*i+1, Range, maxNodes)...)
		}
		if 2*i+2 >= maxNodes {
			return vals
		} else {
			if Range[1] > tree[2*i+2].Range[0] {
				vals = append(vals, traverseBottom(tree, 2*i+2, Range, maxNodes)...)
			}

		}
		return vals
	}

}

// Recursive function that finds the lowest index in the tree list that encapsulates the given Range [a,b] and returns the index
func traverseNearest(tree []*Node, i int, Range [2]float32, maxNodes int) int {

	if 2*i+1 >= maxNodes {
		return i
	} else {
		if Range[0] >= tree[2*i+1].Range[0] && Range[1] <= tree[2*i+1].Range[1] {
			return traverseNearest(tree, 2*i+1, Range, maxNodes)
		} else if 2*i+2 >= maxNodes {
			return i
		} else if Range[0] >= tree[2*i+2].Range[0] && Range[1] <= tree[2*i+2].Range[1] {
			return traverseNearest(tree, 2*i+2, Range, maxNodes)
		} else {
			return i
		}
	}
}

// recursive function that adds appropriate ranges for a node and its children
func build(tree []*Node, newRange [2]float32, i int, maxlen int) {

	if i >= maxlen {
		return
	} else {
		tree[i] = &Node{Range: newRange, Value: 0}
		middle := newRange[0] + (newRange[1]-newRange[0])/2

		build(tree, [2]float32{newRange[0], middle}, 2*i+1, maxlen)
		build(tree, [2]float32{middle, newRange[1]}, 2*i+2, maxlen)

	}

}

// recursive function that increments the Value attribute inside nodes that the parameter value belongs to.
func populate(tree []*Node, i int, value float32, maxlen int) {

	tree[i].Value = tree[i].Value + 1

	if 2*i+1 >= maxlen {
		return
	} else {
		if value < (tree[i].Range[0] + (tree[i].Range[1]-tree[i].Range[0])/2) {
			populate(tree, 2*i+1, value, maxlen)
		} else if 2*i+2 >= maxlen {
			return
		} else {
			populate(tree, 2*i+2, value, maxlen)
		}
	}

}

// This one is the general function to generate the binary tree, with a given dataset, and number of leaves.
func buildTree(leaves int, file string) []*Node {

	nodes := 2*leaves - 1

	dataset, _ := readFile(file)

	treeArray := make([]*Node, nodes)

	min, max := findMinMax(dataset)
	interval := max - min
	middle := min + interval/2
	treeArray[0] = &Node{Range: [2]float32{min, max}, Value: 0}
	build(treeArray, [2]float32{min, middle}, 1, nodes)
	build(treeArray, [2]float32{middle, max}, 2, nodes)

	for _, data := range dataset {
		populate(treeArray, 0, data, nodes)
	}

	return treeArray

}

// this function runs similar timer test to the ones in the article, with a given number of leaves and dataset size.
// the boolean parameters are just to print correct things in correct order when this function is called repeatedly.
// this function returns the tree generation time, time to add noise, key generation time and encryption times.
//
// additionally if dec is true:
//
//		returns the time it took to decrypt with two different query results over 10 rounds of queries,
//
//	 if these times weren't taken, due to treesOnly and dec parameters, then Nil is returned.
func runTests(leaves int, size int, treesOnly bool, dec bool, rounds int) ([][]time.Duration, []time.Duration) {

	//generate tree
	totalStart := time.Now()
	start := time.Now()
	tree := buildTree(leaves, fmt.Sprintf("data%d.txt", size))
	t := time.Now()
	treeGenTime := t.Sub(start)
	if treesOnly {

		//fmt.Printf("%-12d | %-16d | %s\n", size, leaves, treeGenTime)
		return nil, []time.Duration{treeGenTime}
	}

	// Parameters for Laplace noise
	start = time.Now()
	epsilon := 0.5 // didn't find any indication in the paper as to what security parameter value was used so, I came up with this through testing
	µ := float64(0.0)
	b := 1 / (epsilon / math.Log(float64(len(tree)))) // using the 1/epsilon' mentioned in the paper

	AddLaplaceNoise(tree, µ, b)
	t = time.Now()
	addNoiseTime := t.Sub(start)

	//encrypt nodes
	ciphers, msk, scheme, keyGenTime, encryptionTime := encrypt(tree, len(tree), size)

	t = time.Now()
	totalTime := t.Sub(totalStart)

	//fmt.Printf("%-12d | %-15d | %-15s | %-15s | %-14s | %-10s | %s\n", size, leaves*2-1, treeGenTime, addNoiseTime, keyGenTime, encryptionTime, totalTime)

	if !dec {
		return nil, []time.Duration{treeGenTime, addNoiseTime, keyGenTime, encryptionTime, totalTime}
	}

	decryptionTimes := [][]time.Duration{}

	// decrypt 10 random queries
	for i := 0; i < rounds; i++ {

		timescache := []time.Duration{}

		start = time.Now()
		bottom, single := query(tree)
		t = time.Now()
		queryTime := t.Sub(start)

		_, keyGenB, bTime := decrypt(bottom, ciphers, scheme, msk)

		_, keyGenS, sTime := decrypt(single, ciphers, scheme, msk)

		timescache = append(timescache, queryTime, keyGenB, bTime, keyGenS, sTime)
		decryptionTimes = append(decryptionTimes, timescache)
	}

	return decryptionTimes, []time.Duration{treeGenTime, addNoiseTime, keyGenTime, encryptionTime, totalTime}
}

// main function that is mostly for printing and calling the runtests function.
// the leaves and datasizes can however be modified, but there needs to exist a corresponding dataset if datasizes is changed.
func main() {

	rounds, _ := strconv.Atoi(os.Args[2])
	cmd := strings.ToLower(os.Args[1])
	leaves := []int{32, 64, 128, 256, 512, 1024}
	datasizes := []int{100, 500, 1000, 10000}

	if cmd == "tree" || cmd == "treegen" {
		testTreeGen(datasizes, leaves, rounds)
	} else if cmd == "enc" || cmd == "enryption" {
		testEnc(leaves, rounds, 10000)
	} else if cmd == "dec" || cmd == "decryption" {
		testDec(leaves, rounds, 10000)
	} else {
		fmt.Println("Invalid parameters, give in following format: [command] [rounds]\ncommand is tree,enc or dec\nrounds is an integer")
	}

}

// runs testing for tree generation and population
func testTreeGen(datasizes []int, leaves []int, rounds int) {

	fmt.Println("Dataset Size | Number Of Leaves | Average Time")
	for _, data := range datasizes {

		for _, leaf := range leaves {

			var totalTime time.Duration

			for i := 0; i < rounds; i++ {

				_, times := runTests(leaf, data, true, false, 0)
				totalTime += times[0]

			}

			avgTime := time.Duration(int64(totalTime) / int64(rounds))
			fmt.Printf("%-12d | %-16d | %s\n", data, leaf, avgTime)

		}
		fmt.Println("-------------+------------------+--------------")
	}
}

// runs testing for tree generation, noise addition, keygen and encryption times
func testEnc(leaves []int, rounds int, datasize int) {

	fmt.Println("Dataset Size | Number Of Nodes | Tree Generation | Laplacian Noise | Key Generation | Encryption Time | Total Time ")
	for _, leaf := range leaves {

		var totalTreeGen, totalNoise, totalKeyGen, totalEnc, totalTime time.Duration

		for i := 0; i < rounds; i++ {
			_, times := runTests(leaf, datasize, false, false, 0)
			totalTreeGen += times[0]
			totalNoise += times[1]
			totalKeyGen += times[2]
			totalEnc += times[3]
			totalTime += times[4]
		}

		fmt.Printf("%-12d | %-15d | %-15s | %-15s | %-14s | %-15s | %s\n",
			10000,
			leaf*2-1,
			time.Duration(int64(totalTreeGen)/int64(rounds)),
			time.Duration(int64(totalNoise)/int64(rounds)),
			time.Duration(int64(totalKeyGen)/int64(rounds)),
			time.Duration(int64(totalEnc)/int64(rounds)),
			time.Duration(int64(totalTime)/int64(rounds)))

	}
}

// runs testing for QueryTimes, Functional Decryption key Generation and Decryption time
// for both types of queries, with the given dataset.
func testDec(leaves []int, rounds int, datasize int) {

	fmt.Println("Number Of Nodes | Query Time | KeyGen Fast | Decryption Fast | KeyGen Accurate | Decryption Accurate ")
	for _, leaf := range leaves {

		var totalQ, totalKB, totalKS, totalDB, totalDS time.Duration

		decryptionTimes, _ := runTests(leaf, datasize, false, true, rounds)

		for _, times := range decryptionTimes {
			totalQ += times[0]
			totalKB += times[1]
			totalDB += times[2]
			totalKS += times[3]
			totalDS += times[4]
		}

		fmt.Printf("%-15d | %-10s | %-11s | %-15s | %-15s | %-15s\n",
			leaf*2-1,
			time.Duration(int64(totalQ)/int64(rounds)),
			time.Duration(int64(totalKS)/int64(rounds)),
			time.Duration(int64(totalDS)/int64(rounds)),
			time.Duration(int64(totalKB)/int64(rounds)),
			time.Duration(int64(totalDB)/int64(rounds)),
		)

	}

}

// This function encrypts all the nodes of a given binary tree using gofe DDHMulti scheme and times the operations.
func encrypt(tree []*Node, numOfClients int, size int) ([]data.Vector, *simple.DDHMultiSecKey, *simple.DDHMulti, time.Duration, time.Duration) {

	l := 1                               // size of the input vectors per encryptor. 1 since all encryptors have single value
	bound := big.NewInt(int64(size * 2)) // maximum value that can be decrypted in the end.
	modulusLength := 1024

	values := make([]data.Vector, len(tree))

	for i := range values {
		values[i] = []*big.Int{big.NewInt(int64(tree[i].Value))}
	}

	x, _ := data.NewMatrix(values)

	scheme, _ := simple.NewDDHMultiPrecomp(numOfClients, l, modulusLength, bound)
	start := time.Now()
	mpk, msk, _ := scheme.GenerateMasterKeys() //generate keys, for numOfClients number of encryptors.
	t := time.Now()
	keyGenTime := t.Sub(start)

	start = time.Now()
	encryptors := make([]*simple.DDHMultiClient, numOfClients)

	for i := range encryptors {

		encryptors[i] = simple.NewDDHMultiClient(scheme.Params) //create encryptors
	}

	ciphers := make([]data.Vector, numOfClients)

	for i := range encryptors {

		ciphers[i], _ = encryptors[i].Encrypt(x[i], mpk[i], msk.OtpKey[i]) // each encryptor encrypts their plaintextvalue with their public key
	}
	t = time.Now()
	encryptionTime := t.Sub(start)

	return ciphers, msk, scheme, keyGenTime, encryptionTime

}

// This function decrypts the ciphers that are found at indices given as a parameter.
// the sum of values in range [a,b] can be retrieved by computing innerproduct of the indices by y matrix consisting only of 1's
func decrypt(indices []int, ciphers []data.Vector, scheme *simple.DDHMulti, msk *simple.DDHMultiSecKey) (*big.Int, time.Duration, time.Duration) {

	numOfClients := len(indices) // num of clients is the number of nodes that are included in the calculation instead of all nodes.

	dec := simple.NewDDHMultiFromParams(numOfClients, scheme.Params)

	sks := make([]data.Vector, numOfClients)
	otps := make([]data.Vector, numOfClients)
	prunedCiphers := make([]data.Vector, numOfClients)

	// here the appropriate subkeys and ciphers are extracted.
	for i := range indices {

		sks[i] = msk.Msk[indices[i]]
		otps[i] = msk.OtpKey[indices[i]]
		prunedCiphers[i] = ciphers[indices[i]]
	}
	sk, _ := data.NewMatrix(sks)
	otp, _ := data.NewMatrix(otps)

	partialmsk := &simple.DDHMultiSecKey{Msk: sk, OtpKey: otp} // need to create a new key with approrpriate subkeys to be able to decrypt

	yvec := make([]data.Vector, numOfClients)

	for i := range yvec {
		yvec[i] = []*big.Int{big.NewInt(1)}
	}

	y, _ := data.NewMatrix(yvec)

	start := time.Now()
	feKey, _ := dec.DeriveKey(partialmsk, y) // The functional key is created here.
	t := time.Now()
	keyGenTime := t.Sub(start)

	start = time.Now()
	xy, _ := dec.Decrypt(prunedCiphers, feKey, y) // here the inner product of the wanted ciphers and y is calculated with the functional key
	t = time.Now()
	decryptionTime := t.Sub(start)
	return xy, keyGenTime, decryptionTime

}
