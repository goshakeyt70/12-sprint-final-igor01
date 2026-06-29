package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource — это источник псевдослучайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix-формате в виде числа
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
	// (оставить) 1. Инициализация изолированнй БД SQLite в оперативной памяти (:memory:)
	// для ускорения тестов и независимости от файлов на диске.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}
	defer db.Close()

	// ========================================================================
	// Автоматическая подготовока структуры БД СУБД внутри ТЕСТА.
	// ========================================================================

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TEXT NOT NULL	
	);`

	// Принудительно создаем таблицу перед тем, как выполнить тесты: ADD, GET или DELETE.
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Ошибка подготовки таблицы parcel для тестов: %v", err)
	}

	// Очищаем таблицу от мусора прошлых запусков.
	_, _ = db.Exec("DELETE FROM parcel;")
	_, _ = db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'parcel';")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// get
	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	require.Equal(t, parcel, stored)

	// delete
	err = store.Delete(parcel.Number)

	stored, err = store.Get(parcel.Number)
	require.Equal(t, sql.ErrNoRows, err)
}

// ================================================ Следующий тест: ADD, SET, CHECK adress
// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// - 2-ое создание БД
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}
	defer db.Close()
	//-
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TEXT NOT NULL	
	);`
	//-
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Ошибка подготовки таблицы parcel для тестов: %v", err)
	}
	//-
	_, _ = db.Exec("DELETE FROM parcel;")
	_, _ = db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'parcel';")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)

	require.NoError(t, err)

	// check
	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	require.Equal(t, newAddress, stored.Address)
}

// ================================================== Следующий тест: ADD, SET STATUS, CHECK
// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// - 3-ие создание БД
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}
	defer db.Close()
	//-
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TEXT NOT NULL	
	);`
	//-
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Ошибка подготовки таблицы parcel для тестов: %v", err)
	}
	//-
	_, _ = db.Exec("DELETE FROM parcel;")
	_, _ = db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'parcel';")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set status
	err = store.SetStatus(parcel.Number, ParcelStatusSent)

	require.NoError(t, err)

	// check
	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, stored.Status)
}

// ======================================================== Следующий тест: ADD, GET by CLIENT, CHECK
// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// - 4-ое создание БД
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}
	defer db.Close()
	//-
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TEXT NOT NULL	
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Ошибка подготовки таблицы parcel для тестов: %v", err)
	}
	//-
	_, _ = db.Exec("DELETE FROM parcel;")
	_, _ = db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'parcel';")

	/*
		// prepare
		db, err := sql.Open("sqlite", "tracker.db")
		if err != nil {
			require.NoError(t, err)
		}
		defer db.Close()
	*/

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам одного клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])

		require.NoError(t, err)
		require.NotEmpty(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)

	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]

		require.True(t, ok)
		require.Equal(t, expectedParcel, parcel)
	}
}
