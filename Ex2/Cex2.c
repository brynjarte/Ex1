
#include <pthread.h>
#include <stdio.h> 

int i = 0;
pthread_mutex_t mtx;

void* thread_1Func() {
	int j;
	
	
	for (j = 0; j < 1000000; j++) {
	pthread_mutex_lock(&mtx);
		i++;
	pthread_mutex_unlock(&mtx);
	}
	
	return NULL;
}

void* thread_2Func() {
	int k;
	
	for (k = 0; k < 999999; k++) {
	pthread_mutex_lock(&mtx);
		i--;
	pthread_mutex_unlock(&mtx);
	}
	
	return NULL;
}


int main(){
	pthread_mutex_init(&mtx,NULL);

	pthread_t thread_1;
	pthread_t thread_2;
	
	

	
	pthread_create(&thread_1, NULL, thread_1Func, NULL);
	pthread_create(&thread_2, NULL, thread_2Func, NULL);
	pthread_join(thread_1, NULL);
	pthread_join(thread_2, NULL);
	
	
	pthread_mutex_destroy(&mtx);
	printf("%d\n", i);

	return 0;
}
