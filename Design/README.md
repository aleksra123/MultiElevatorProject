# Design review
### Strengths and weaknesses of programming language Go
##### Strengths
* First class support for concurrency, as it is a built-in feature. By using `goroutines` you can run several lightweight threads and therefore have multiple processes running simultaneously and efficiently. This will be very useful in the elevator project as we have some fast and some slow processes. Polling, calculations, network communication are processes using different amounts of time and being able to do all simultaneously is our number one reason for using Go. `channels` is an efficient way of communicating between `goroutines`
* It is very fast, which is a good thing. Go  
##### Weaknesses
* Go is a fairly young language, which means that you can't find as much help online as you would with older and more common languages. E.g. Python, C, C++. 

##### Both
* By using Go, we have to use UDP (User Datagram Protocol) for communcationa between the nodes. It is easy to use, because it is minimal and connectionless. UDP is unreliable, which means that our code must be fault tolerant. The messages can be either corrupted or lost in transmission. 
### Abstraction in Go
New language for us, but fairly known syntax. We have used C, C++ and Python earlier. 
