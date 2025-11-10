package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
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
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	added, err := store.Add(parcel)

	require.NoError(t, err)
	assert.NotEmpty(t, added)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	got, err := store.Get(added)

	require.NoError(t, err)
	assert.Equal(t, parcel.Client, got.Client)
	assert.Equal(t, parcel.Status, got.Status)
	assert.Equal(t, parcel.Address, got.Address)
	assert.Equal(t, parcel.CreatedAt, got.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(added)
	require.NoError(t, err)

	_, err = store.Get(added)
	assert.ErrorContains(t, err, "not found")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	added, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, added)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки

	newAddress := "new test address"
	err = store.SetAddress(added, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	got, err := store.Get(added)
	require.NoError(t, err)
	assert.Equal(t, newAddress, got.Address)
	assert.Equal(t, parcel.Client, got.Client)
	assert.Equal(t, parcel.Status, got.Status)
	assert.Equal(t, parcel.CreatedAt, got.CreatedAt)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	added, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, added)
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(added, ParcelStatusSent)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	got, err := store.Get(added)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, got.Status)
	assert.Equal(t, parcel.Client, got.Client)
	assert.Equal(t, parcel.Address, got.Address)
	assert.Equal(t, parcel.CreatedAt, got.CreatedAt)
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
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}
	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Equal(t, len(parcels), len(storedParcels))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		got, ok := parcelMap[parcel.Number]
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.True(t, ok, parcel)
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, got.Client, parcel.Client)
		assert.Equal(t, got.Status, parcel.Status)
		assert.Equal(t, got.Address, parcel.Address)
		assert.Equal(t, got.Number, parcel.Number)
		assert.Equal(t, got.CreatedAt, parcel.CreatedAt)
	}
}
