from threading import Thread
from threading import Lock

i = 0
mtx = Lock()

def thread_1Func():
	global i
	j = 0
	
	while j < 1000000:
		mtx.acquire()
		i = i+1
		j = j+1
		mtx.release()

def thread_2Func():
	global i
	j = 0
	
	while j < 999999:
		mtx.acquire()
		i = i-1
		j = j+1
		mtx.release()
		
def main():

	thread_1 = Thread(target = thread_1Func, args = (), )
	thread_1.start()
	thread_1.join()

	thread_2 = Thread(target = thread_2Func, args = (), )
	thread_2.start()
	
	thread_2.join()
	
	print i


main()
