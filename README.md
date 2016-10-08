# It is not a crawler!

It just implements a crawler dispatcher, which can fetch url no repeatly.  
A map is used to save the fetched url among many gorountines.  
But the map doesn't use sync.Mutex.  
Just use channel.  

If you want know more, just read the code.  
__StateProcessor__ is used to update map and dispatch.  
__Crawl__ is used to communicate with __StateProcessor__.   
__IdelMonitor__ is used to update number of idle crawlers secondly. When all crawler idle, the whole program exits.  

## This is a no_buffer_chan branch.
I think when you use channel, you should gurantee gorountines doesn't deadlock with no buffer channel.  
And this will gurantee it works with buffered channel.  

When you use channel, you can just use no buffer channel to debug and ensure that there will be no deadlock.  
Then change it to buffered channel to improve multi-gorountines' performance.     
