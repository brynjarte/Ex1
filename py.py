from threading import Thread

i = 0

def thread_1Func():
	global i
	j = 0
	while j < 1000000:
		i = i+1
		j = j+1

def thread_2Func():
	global i
	j = 0
	while j < 1000000:
		i = i-1
		j = j+1
		
def main():
	thread_1 = Thread(target = thread_1Func, args = (), )
	thread_1.start()
	thread_1.join()

	thread_2 = Thread(target = thread_2Func, args = (), )
	thread_2.start()
	
	thread_2.join()
	
	print i


main()
