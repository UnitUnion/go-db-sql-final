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
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	checkParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, parcel, checkParcel)
	err = store.Delete(parcel.Number)
	require.NoError(t, err)
	_, err = store.Get(parcel.Number)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	defer db.Close()
	parcel := getTestParcel()
	require.NoError(t, err)
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)
	checkParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, checkParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	defer db.Close()
	parcel := getTestParcel() // получение тестовой посылки
	require.NoError(t, err)   // проверка ошибки
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	err = store.SetStatus(parcel.Number, ParcelStatusDelivered)
	require.NoError(t, err)
	checkParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusDelivered, checkParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))
	for _, parcel := range storedParcels {
		require.Contains(t, parcelMap, parcel.Number)
		require.Equal(t, parcelMap[parcel.Number], parcel)
	}
}
