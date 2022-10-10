package main

import "testing"

// unit test for pathExists function
func TestPathExists(t *testing.T) {

	// test path exists
	stat, fileExists, err := pathExists("main.go")

	if err != nil {
		t.Errorf("Error checking path: %v", err)
	}

	if !fileExists {
		t.Errorf("Path should exist")
	}

	if stat.IsDir() {
		t.Errorf("Path should be a file")
	}

	// test path does not exist
	stat, fileNonExist, err := pathExists("nonsexist.go")

	if err == nil {
		t.Errorf("Drop Error in case file not exists")
	}

	if fileNonExist {
		t.Errorf("Path should not exist")
	}

	if stat != nil {
		t.Errorf("Stat should be nil")
	}
}

// unit test for readJSON function
func TestReadJSON(t *testing.T) {

	// test read json file
	res := readJSON("test/parameters.json", "dev")

	if res["REACT_APP_PARAM1"] != "VALUE1" {
		t.Errorf("REACT_APP_PARAM1 value should be VALUE1")
	}

	if res["REACT_APP_PARAM2"] != "VALUE2" {
		t.Errorf("REACT_APP_PARAM2 value should be VALUE2")
	}

	// // test read json file with wrong key
	// res = readJSON("test/parameters.json", "qa")

	// if len(res) != 0 {
	// 	t.Errorf("Result should be empty")
	// }

	// // test read json file with wrong file
	// res = readJSON("test1.json", "test")

	// if len(res) != 0 {
	// 	t.Errorf("Result should be empty")
	// }
}
