//go:build !solution

package main

import (
	"sort"
	"sync"
)

type KeyLock struct {
	locks   map[string]chan struct{}
	mapLock sync.Mutex
}

func New() *KeyLock {
	return &KeyLock{locks: make(map[string]chan struct{})}
}

func (l *KeyLock) LockKeys(Keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	// один из тестов смотрит на то чтобы Keys не менялся, поэтому делаю копию для работы с ней
	keys := make([]string, len(Keys))
	copy(keys, Keys)
	// может быть такой кейс gorutine1 пришла с {a,b} gorutine2 {b,a} и мы можем получить deadlock
	sort.Strings(keys)
	l.mapLock.Lock()
	// проходка по ключам, и проверка. Если ключ отсутствует то добавляем канал с локом для него
	for _, k := range keys {
		if _, ok := l.locks[k]; !ok {
			l.locks[k] = make(chan struct{}, 1)
			l.locks[k] <- struct{}{}
		}
	}
	l.mapLock.Unlock()

	capturedKeys := make([]string, 0)
	// проходим по всем ключам и пытаемся захватить каждый лок
	for _, k := range keys {
		// мапа не потокобезопасна поэтому необходима защита через mutex
		l.mapLock.Lock()
		ch := l.locks[k]
		l.mapLock.Unlock()

		select {
		// берем токен из канала. Если получилось то ключ считается захваченным
		case <-ch:
			capturedKeys = append(capturedKeys, k)
		case <-cancel:
			// если пришел cancel то освобождаем все захваченные ключи
			l.mapLock.Lock()
			for _, k := range capturedKeys {
				l.locks[k] <- struct{}{}
			}
			l.mapLock.Unlock()
			return true, nil
		}
	}
	// функция для разблокировки всех захваченных ключей
	unlock = func() {
		l.mapLock.Lock()
		defer l.mapLock.Unlock()

		for _, k := range keys {
			l.locks[k] <- struct{}{}
		}
	}

	return false, unlock
}
