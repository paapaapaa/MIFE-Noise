# Multi-input Functional Encryption with Laplacian noise for differential privacy


## Context
This Project was an assignment for the course Security Protocols.  
The aim of the program is to replicate the experiments detailed in section 9 of the paper [Private Lives Matter: A Differential Private Functional Encryption Scheme by Alexandtros Bakas, Antonis Michalas, and Tassos Dimitriou](https://eprint.iacr.org/2021/1692). My implementation heavily relies on the [GoFe](https://github.com/fentec-project/gofe) functional encryption library for Go, which was used to instantiate a similar scheme to the one used in the paper. The aim of the experiments is to prove that using differential privacy on top of encryption does not add any significant time constraints to the overall runtime of the scheme. When creating my program I tried to follow the experiments closely and indeed my results were very similar to the ones in the paper. 

The most interesting parts of the program are the functions AddLaplaceNoise, Which adds noise to the data to obscure the individual impact of an encryptor to the test results, Encrypt, which generates the simulated encryptors and their corresponding encryption keys, and Decrypt, which derives the functional decryption keys for the given vector of encrypted data. The other parts of the program were hastily created off the top of my head as the optimization of the tree structure and testing was not necessary for the experiment.

## Prerequisites
 - GoFe functional encryption library and its dependencies found [here](https://github.com/fentec-project/gofe)
 - The other imports should be native to Go so anything else shouldn't be needed to run.
 - There might be a need to change values or generate new instances of go.mod and/or go.sum, which I provided with the project, I'm not very experienced with Go, so I don't really know.
   
## Usage

The application requires two command line parameters, the first one being a command either tree,enc or dec. Each of them produces tables similar to the ones detailed in the paper. The second parameter is an integer of how many iterations do you want the program to compute the averages over.  

an example execution would be:  
- go run cw2.go tree 50
  
this would provide time averages for the binary tree generation over 50 rounds.  

the program can also be built first with  
- go build cw2.go and then ran with  
- cw2.exe tree 50 on windows  

the round numbers I would suggest are:  
- go run cw2.go tree 50  
- go run cw2.go enc 10  
- go run cw2.go  5
  
the latter two take considerably longer than the first.  

## Results

My results were very similar to the ones described in the paper, which is that introducing differential privacy takes no time at all.  
More detailed version of my analysis that was part of the assignment can be found attached as a [pdf](Analysis.pdf)



## Contact Information

- Joona Korkeamäki
- joona.v.korkeamaki@tuni.fi
