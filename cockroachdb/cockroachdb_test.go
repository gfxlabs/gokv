package cockroachdb_test

import (
	"database/sql"
	"log"
	"testing"

	"github.com/philippgille/gokv/cockroachdb"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to CockroachDB works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to CockroachDB could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to CockroachDB works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to CockroachDB could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the CockroachDB client.
//
// Note: This test is only executed if the initial connection to CockroachDB works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to CockroachDB could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to CockroachDB works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to CockroachDB could be established. Probably not running in a proper test environment.")
	}

	// Test empty key
	client := createClient(t, encoding.JSON)
	err := client.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = client.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = client.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
//
// Note: This test is only executed if the initial connection to CockroachDB works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to CockroachDB could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, encoding.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, encoding.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, codec)

			// Prep
			err := client.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = client.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
			_, err = client.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = client.Get("foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(encoding.JSON))
	t.Run("get with nil / nil value parameter", createTest(encoding.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to CockroachDB works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to CockroachDB could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, encoding.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	db, err := sql.Open("postgres", "postgres://root@localhost:26257/?sslmode=disable")
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}

	err = db.Ping()
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}

	return true
}

func createClient(t *testing.T, codec encoding.Codec) cockroachdb.Client {
	options := cockroachdb.Options{
		Codec: codec,
		// Higher values seem to lead to issues on Travis CI when using MySQL,
		// so let's just use the same value here.
		MaxOpenConnections: 25,
	}
	client, err := cockroachdb.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}