package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // Подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel) // Добавляем посылку
	require.NoError(t, err)
	defer db.Close() // Закрыть соединение

	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	storedParcel, err := store.Get(id) // Получаем посылку
	require.NoError(t, err)
	require.Equal(t, parcel.Status, storedParcel.Status)   // Проверяем статус
	require.Equal(t, parcel.Address, storedParcel.Address) // Проверяем адрес
	require.Equal(t, parcel.Client, storedParcel.Client)   // Проверяем клиента

	// delete
	err = store.Delete(id) // Удаляем посылку
	require.NoError(t, err)

	// Проверка на отсутствие
	_, err = store.Get(id) // Проверяем, что нельзя получить удалённую посылку
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress) // Обновляем адрес
	require.NoError(t, err)

	// check
	storedParcel, err := store.Get(id) // Получаем обновлённую посылку
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address) // Проверяем новый адрес
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set status
	err = store.SetStatus(id, ParcelStatusSent) // Обновляем статус
	require.NoError(t, err)

	// check
	storedParcel, err := store.Get(id) // Получаем посылку
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, storedParcel.Status) // Проверяем статус
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client       // Задаем всем посылкам один и тот же идентификатор клиента
		id, err := store.Add(parcels[i]) // Добавляем посылку
		require.NoError(t, err)
		parcels[i].Number = id
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // Получаем все посылки по идентификатору клиента
	require.NoError(t, err)
	require.Equal(t, 3, len(storedParcels)) // Проверяем количество

	// check
	for _, parcel := range storedParcels {
		require.Contains(t, parcels, parcel) // Проверяем соответствие посылок
	}
}
