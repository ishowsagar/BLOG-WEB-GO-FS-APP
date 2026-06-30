package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
)

// @ rag implemented handlers for resolving rag

// types

// type struct which -> stores actual text & relevant vectorised chunks
type ChunkVector struct {
	Text string
	Chunk []float64 // chunk would be vectorised chunk - in form of embedded float values in slice of float64 
}


// type struct which stores each chunkVector + its score ( sorted from sort.Slice)
type ScoredChunk struct {
	Chunk *ChunkVector
	Score float64

}

// file where slice of scored codebase files vectors are stored
var CodebaseVectorContextFileName = "codebase_vectors.json"

// Rag query model flow

	// 1. takes question from the user with contextChunk  — converted to a vector ( need way to store all codebase into vectorised form)
	// would have all codebase chunks — each converted to a vector and stored in the index.
	//  need to get nearest/similar vector from all chunks stored in []cbVectors 
	// 2. but before 1, need context chunks as from...vague -> By using findTopNChunks func to -> loop over codebase vectors (already indexed) and providing queryVector slice + currentVector slice
	// finding and sorting slice which has all codebase vector with similairty score -> sorting them all in slice from score in descending order
	// 3. Now, all codebase vectors upto N in descending order is stored -> getting only those vectors with highest similarity ( as sorting them by computred similairty for that current vector and since similairty is computed from both slices vectors maths)
	// 4. returning those nearest vectors ->  for context window of querying -> these vectors are determined by query vector

	// but for conversion of question into query vector -> need indexer ( converts go files into vectors)




// func that takes in both query and retrieved chunkVector from codebase -> returns 'n' -> nearest top chunks 
func EvaluateCosineSimilarity(queryVector,relevantChunkVector []float64)float64 {


	// loop over vectors el to get dot product and magnitudes
	var( dotProduct float64;
		queryVectorMagnitude float64;
		relevantChunkVectorMagnitude float64 
	)

	// visualisation -> IF queryV [0.1,0.2 ...], relevantVector [0.4,0.3...]
	// we loop on qV slice of float64 el -> for each at that position -> take vector of that slice * other at same position as {... i keep the val same, just puls i as currewn titeration}

	// then for each current iteration's we get that el from position -> sum accordingly
	// ! it's unidiomatic to have blank 'value' indentifier for val identifier
	for i:= range queryVector {

		// ** 'i' is just for current iteration index -> to get value of that positioned element in slices

		// ** dot product is each slice's current indexed positioned elements (a.k.a el) dot product's -> accumulated sum
		dotProduct += queryVector[i] * relevantChunkVector[i] // sum of eachIndex of q and r vector indexes -> queryVector current position's vector el into rele's vector
		
		// ** magnitude is each's slice current (corresponding) el sqred against itself
		// sum of each slice's current index's element (floating num value) -> accumulated sum
		queryVectorMagnitude += queryVector[i] * queryVector[i] // each index's sq sum of queryVector
		relevantChunkVectorMagnitude += relevantChunkVector[i] * relevantChunkVector[i] // sum of each chunk's vector's  current index -> 0.1(a) * 0.2(b) + ... + 

	}

	// this formula might need to learn more about -> not formula but clicky feel why this only

	// finds N as most similar chunks

	// this takes sum of correspoding slice's each el dot product accumulated sum + sqred accumulated sum product in division -> gives top n chunks
	similarity := dotProduct/(math.Sqrt(queryVectorMagnitude)*math.Sqrt(relevantChunkVectorMagnitude))
	return similarity 

}



// this func (a.k.a  function) takes in queryVector and codebaseVectors and n chunks -> if found -> return those chunks slice
func FindNearestTopNchunks(codebaseVectors []*ChunkVector,queryVector []float64,topN int) ( relevantChunksVectors []*ScoredChunk) {

	// named returns actually intializes var -> directly return from there -> so it comes with this purpose too, i thought just for clarification purpose of returns

	// testing - returning this whole slice -> it has both score and needed chunkVector both

	// tracks each vector with similarity score
	var scoredVectors []*ScoredChunk

	// need to iterate over each codebaseVectors -> store similairty and related chunk in []*ScoredChunks which <- stores both
	for _,currentChunkVector := range codebaseVectors{
		//  compute similarity for each chunk ->
		evaluatedSimilarityNum := EvaluateCosineSimilarity(queryVector,currentChunkVector.Chunk)

		// store computed similarity in struct with current chunk vector - ....ohhh this tracks both anaylysed current vector and its similarity score, so we then get those which we need based off score
		scoredChunk := &ScoredChunk{
			Chunk: currentChunkVector,
			Score: evaluatedSimilarityNum,
		}

		// need to store these scoredChunks -> whcih stores these scoredChunks els
		scoredVectors = append(scoredVectors, scoredChunk)
	}
	
	// ! bug - sorting it out after full slice is obtained by <- appending those el which have computed score as similarity and its related chunk
	// sort's el of slice in desc order --> to get highest first
	sort.Slice(scoredVectors,func (i,j int) bool {
		//  so for every current el -> it keep sorting untill it does not sort in way -> highest degree (degree say for val now) comes firsr, return those
		return scoredVectors[i].Score > scoredVectors[j].Score //descending ordered el
	})

	// guard check - we don't wanna slice it out till that positioned el which is not in slice already
	if topN > len(scoredVectors) {
		topN = len(scoredVectors) // set topN to be same as max length of the slice
	}
	//  slice sorted by score obtained -> return those but upto first n desired
	relevantChunksVectors = scoredVectors[:topN] // slice of slice till :N position only like we did with read bytes -> slice upto that read bytes only
	
	return relevantChunksVectors
}


// func whose function is to -> take the chunkVector to -> return 'n' as relevant and most similar chunks
func VectorResolver(codebaseVectors []*ChunkVector,quesVector []float64, ) {

	// its job is to return -> most similar "n" chunks -> as n vector -> need to find nearest from this vector

	
	// Need a way to -> get closest vector related to queryVector from all chunks vectors stored in []*chunkVectors -> that resprest that chunk entirely <- purpose
	
	// to check that closed approximation -> need to do some maths

}


// method that belongs to clientConfig type -> returns relevant chunks's embedded vectors
func(cfg *ClientConfig) RagReq(requestUrl,userPrompt string) ([]float64,error){

	// need struct to store chunk with index
	
	textWrapper := vectorOutTextWrapperGem{
					// todo - might we later need prompt build function to return desired prompt
					Text: userPrompt,
				}
	// construct outbound
	outbound:= vectorOutBodyGem{
		Model: "models/gemini-embedding-001",
		Content: &vectorOutContentWrapperGem{
			Parts: []*vectorOutTextWrapperGem{
				&textWrapper,
			},
		},
	}

	
	// marshal out
	  
	outByte,err := json.Marshal(outbound)
	if err != nil {
		return nil,err
	}

	// read bytes into buffer -> buffer satisfies reader/wr both -> reader
	// bug - ahh, read method read into bytes,not read from 
	// fix - new reader source - either write to buf -> then read or totally new reader
	bufReader := bytes.NewReader(outByte) //new ioreader soruce from bytes - when bytes, when str -> use strings.NewR...

	// creating a new request
	req,err := http.NewRequest("POST",requestUrl,bufReader)
	if err != nil {
		return nil,err
	}

	// setting headers
	req.Header.Set("Content-Type","application/json") 
	req.Header.Set("x-goog-api-key",cfg.secretKey)

	// do request
	res,err := cfg.httpClient.Do(req)
	if err != nil {
		return nil,err
	}

	// resolving errors
	if res.StatusCode == 429 {
		errMsg := "\n\033[31m🚨 Rate Limit Exceeded (HTTP 429).\033[0m"
		fmt.Println(errMsg)
		return nil,errors.New(errMsg)
	}

	if res.StatusCode != 200 {
		errMsg := "\n\033[31m🚨 failed to do request(%d).\033[0m"
		fmt.Printf(errMsg,res.StatusCode)
		return nil,errors.New(errMsg)
	} 

	// when res is successfull(200) -> decoding into inbound -> getting values ( res vectors -> relevant vectors for conversion into res ) ....vague
	
	resStream := res.Body //ioreader stream source

	// read bytes from stream
	resBodBytes,err := io.ReadAll(resStream)
	if err != nil {
		// only return custom err when there is err handeling check from no implicit error is there that has to be send to the client
		return nil,err
	}

	// validating if res was ok but no bytes in stream ( nothing in the inbound)
	if len(resBodBytes) == 0 {
		// early return
		return nil,errors.New(Red+"inbound empty payload ⚠️")
	}
	// decode and validate inbound
	var inbound *vectorInboundGem
	
	err = json.Unmarshal(resBodBytes,&inbound)
	if err!= nil {
		return nil,err
	}

	// now res is ok + recieved inbound that is healthy too -> returning vectors from the inbound
	if inbound.Embedding == nil {
		return nil,errors.New("nil payload")
	}
	if inbound.Embedding.Values == nil {
		return nil,errors.New("server sent nill vectors values")
	}

	// if these are not nil and res vector is returned -> fetch it
	resVectors := inbound.Embedding.Values // [] of foat64 type of vector els

	return resVectors,nil

}



//  func that chunks down the contents into dedicated chunks -> for indexing purpose of go codebase <- not perfectly accurate but very doable
func ChunkingContent(contentText string,chunkSize int,overlap int) ([]string) {

	// need a way to break down text into 50 lines chunks - overlapped lines (allowance for bigger neccessary chunks)

	// cntLength := len(contentText)

	// // test - read content into buf 
	// buf := make([]byte,chunkSize)

	// contentIOreader := strings.NewReader(contentText)
	
	// n,err := contentIOreader.Read(buf) // read into buf upto chunkSize for flexible reads
	// if err != nil {
	// 	return nil,err
	// }

	// if cntLength 


	var chunkedContentAccumulator []string

	// split text into lines - "\n" linebreaker
	lines := strings.Split(contentText,"\n")

	// implicit loop 
	for i:=0; i<len(lines); i += chunkSize - overlap {
		// for each iteration - split lines slice from current line to chunksize +i (i -> incremented from there - last)

		// taking out slice from lines as those are broken -> from current line to upto chunkSize setting +i for next iteration

		// append each chunk to the chunk accumulator
		 //joining back slice of strings into string
		
		// if i+chunk len is more than -> lines length -> set it to be eaual lines length
		end :=  i + chunkSize 
		if end > len(lines) {  
			end = len(lines)
		}
		
		chunk := lines[i:end]
		chunkedCnt := strings.Join(chunk,"\n")
		chunkedContentAccumulator = append(chunkedContentAccumulator,chunkedCnt)
		
	}
	return chunkedContentAccumulator
}


// func that takes in files, generate rag chunks vectors returns ->  them in a slice -> builds index {vector context for each file contents} ....ohhh this is actually building context for itself?
func(cfg *ClientConfig) Indexer(reqUrl string,fileContents []string) ([]*ChunkVector,error){


	// all it does is -> take each file -> break into chunks ->  get relevant vectors for each -> construct vector for each resp val + content ...vague

	var chunkVectorsAccumulator []*ChunkVector
	// 1.indexes each file content bvy looping into each file
	for _,currentFileContent := range fileContents{
		// calling chunk content on each current filecontent
		Chunks := ChunkingContent(currentFileContent,50,8) // 50 each chunk size, offset by 8 lines
		
		
		//2.  get each's file broken chunk content's vector from req
		for _,chunk := range Chunks{
			// ohh...for each chunk -> calling RagReq to get its vector
			
			//3. retrieving vector for each chunk
			chunkVectorRes,err := cfg.RagReq(reqUrl,chunk)
			if err != nil {
				return nil,err
			}
			// ques - we are not still sending batch by batch req -> get vectors?????

			// construct chunkVector
			chunkVector := &ChunkVector{
				Text: chunk, // current chunk
				Chunk: chunkVectorRes, // response vector values
			}

			//4. tracking vectors
			chunkVectorsAccumulator = append(chunkVectorsAccumulator, chunkVector)
			// track cVectors and store in slice

			// let loop finishes -> append all the chunkVectord

		}


	}

	
	
	//** need to write this slice to the file -> saving it in disk for now -> later switch to vector db itself
	
	// encode slice
	encodedVectorsSlice,err := json.Marshal(chunkVectorsAccumulator)
	if err != nil {
		return nil,err
	}
	
	
	//write to the file

	// ! os.Open -> just for reads, cannot write but it writes too -> truncates if exists otherwise create it
	file,err :=os.Create(CodebaseVectorContextFileName)
	if err != nil {
		return nil,err
	}
	
	defer file.Close() //deferred close
	_,err = file.Write(encodedVectorsSlice) // since file type satisfies write interface -> it has underlying write method which based on the writer, writes provided bytes to the respective destination 
	if err != nil {
		return nil,err
	}
	
	// return slice of that cVector 
	return chunkVectorsAccumulator,nil

}