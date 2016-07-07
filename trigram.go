package trigram

import "sort"

type Trigram uint32

type docList []int

func (d docList) Len() int           { return len(d) }
func (d docList) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d docList) Less(i, j int) bool { return d[i] < d[j] }

//The trigram indexing result include all Document IDs and its Frequence in that document
type IndexResult struct {
	//Save all trigram mapping docID
	DocIDs map[int]struct{}
}

// Extract one string to trigram list
// Note the Trigram is a uint32 for ascii code
func ExtractStringToTrigram(str string) []Trigram {
	if len(str) == 0 {
		return nil
	}

	var result []Trigram
	for i := 0; i < len(str)-2; i++ {
		var trigram Trigram
		trigram = Trigram(uint32(str[i])<<16 | uint32(str[i+1])<<8 | uint32(str[i+2]))
		result = append(result, trigram)
	}

	return result
}

type TrigramIndex struct {
	//To store all current trigram indexing result
	TrigramMap map[Trigram]IndexResult

	//it represent and document incremental index
	maxDocID int
}

//Create a new trigram indexing
func NewTrigramIndex() *TrigramIndex {
	t := new(TrigramIndex)
	t.TrigramMap = make(map[Trigram]IndexResult)
	return t
}

//Add new document into this trigram index
func (t *TrigramIndex) Add(doc string) int {
	newDocID := t.maxDocID + 1
	for _, tg := range ExtractStringToTrigram(doc) {
        mapRet, exist := t.TrigramMap[tg]
		if !exist {
			t.TrigramMap[tg] = IndexResult{
                map[int]struct{}{newDocID: struct{}{}},
            }
		} else {
			//trigram already exist on this doc
			if _, docExist := mapRet.DocIDs[newDocID]; !docExist {
				mapRet.DocIDs[newDocID] = struct{}{}
			}
            t.TrigramMap[tg] = mapRet
		}
	}

	t.maxDocID = newDocID
	return newDocID
}

//This function help you to intersect two map
func IntersectTwoMap(IDsA, IDsB map[int]struct{}) map[int]struct{} {
	var retIDs map[int]struct{}   //for traversal it is smaller one
	var checkIDs map[int]struct{} //for checking it is bigger one
	if len(IDsA) >= len(IDsB) {
		retIDs = IDsB
		checkIDs = IDsA

	} else {
		retIDs = IDsA
		checkIDs = IDsB
	}

	for id, _ := range retIDs {
		if _, exist := checkIDs[id]; !exist {
			delete(retIDs, id)
		}
	}
	return retIDs
}

//Query a target string to return the doc ID
func (t *TrigramIndex) Query(doc string) docList {
	trigrams := ExtractStringToTrigram(doc)

	//Find first trigram as base for intersect
	retObj, exist := t.TrigramMap[trigrams[0]]
	if !exist {
		return nil
	}
	retIDs := retObj.DocIDs

	//Remove first one and do intersect with other trigram
	trigrams = trigrams[1:]
	for _, tg := range trigrams {
		checkObj, exist := t.TrigramMap[tg]
		if !exist {
			return nil
		}
		checkIDs := checkObj.DocIDs
		retIDs = IntersectTwoMap(retIDs, checkIDs)
	}

	return getMapToSlice(retIDs)

}

//Transfer map to slice for return result
func getMapToSlice(inMap map[int]struct{}) docList {
	var retSlice docList
	for k, _ := range inMap {
		retSlice = append(retSlice, k)
	}
	sort.Sort(retSlice)
	return retSlice
}
