# It is not a crawler!

It just implements a crawler dispatcher, which can fetch url no repeatly.  
A map is used to save the fetched url among many gorountines.  
But the map doesn't use sync.Mutex.  
Just use channel.  

If you want know more, just read the code.  
`StateProcessor()` is used to update map and dispatch.  
`Crawl()` is used to communicate with `StateProcessor()`.   
`IdleMonitor()` is used to update number of idle crawlers secondly.   When all crawler idle, the whole program exits.
