package decide

import (
	"testing"
	"math"
)

func TestRule0(t *testing.T) {
	decide := Decide{}
	points := make([][2]float64, 2)

	input := INPUT{}
	input.NumPoints = 2
	input.Parameters.LENGTH1 = 10
	points[0] = [2]float64{0, 0}
	points[1] = [2]float64{0, 5}
	input.Points = points

	decide.input = input



	v, err := decide.Rule0()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}

	input.Parameters.LENGTH1 = 3
	decide.input = input

	v, err = decide.Rule0()
	if err != nil {
		t.Error(err)
		return
	}
	if !v {
		t.Error("Expected true")
		return
	}
}

func TestRule1(t *testing.T) {
	decide := Decide{}
	points := make([][2]float64, 3)

	input := INPUT{}
	input.NumPoints = 3
	input.Parameters.RADIUS1 = 5
	points[0] = [2]float64{5, 0}
	points[1] = [2]float64{0, 5}
	points[2] = [2]float64{5, 5}
	input.Points = points

	decide.input = input

	v, err := decide.Rule1()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}

	input.Parameters.RADIUS1 = 2
	decide.input = input

	v, err = decide.Rule1()
	if err != nil {
		t.Error(err)
		return
	}
	if !v {
		t.Error("Expected true")
		return
	}
}

func TestRule2(t *testing.T) {
	decide := Decide{}
	points := make([][2]float64, 3)

	input := INPUT{}
	input.NumPoints = 3
	input.Parameters.EPSILON = math.Pi * 3/ 4
	points[0] = [2]float64{1, 0}
	points[1] = [2]float64{0, 0}
	points[2] = [2]float64{0, 1}
	input.Points = points

	decide.input = input

	v, err := decide.Rule2()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}

	input.Parameters.EPSILON = 0
	decide.input = input

	v, err = decide.Rule2()
	if err != nil {
		t.Error(err)
		return
	}
	if !v {
		t.Error("Expected true")
		return
	}

	points[0] = [2]float64{0, 0}
	points[1] = [2]float64{0, 0}
	points[2] = [2]float64{0, 1}
	input.Points = points

	decide.input = input

	v, err = decide.Rule2()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}
}

func TestRule3(t *testing.T) {
	decide := Decide{}
	points := make([][2]float64, 3)

	input := INPUT{}
	input.NumPoints = 3
	input.Parameters.AREA1 = 5
	points[0] = [2]float64{5, 0}
	points[1] = [2]float64{0, 2}
	points[2] = [2]float64{0, 0}
	input.Points = points

	decide.input = input

	v, err := decide.Rule3()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}

	input.Parameters.AREA1 = 4.9
	decide.input = input

	v, err = decide.Rule3()
	if err != nil {
		t.Error(err)
		return
	}
	if !v {
		t.Error("Expected true")
		return
	}
}

func TestRule4(t *testing.T) {
	decide := Decide{}
	points := make([][2]float64, 3)

	input := INPUT{}
	input.NumPoints = 3
	input.Parameters.Q_PTS = 3
	input.Parameters.QUADS = 2

	points[0] = [2]float64{1, 1}
	points[1] = [2]float64{-1, 1}
	points[2] = [2]float64{1, 1}
	input.Points = points

	decide.input = input

	v, err := decide.Rule4()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}

	points[2] = [2]float64{1, -1}
	input.Points = points
	decide.input = input

	v, err = decide.Rule4()
	if err != nil {
		t.Error(err)
		return
	}
	if !v {
		t.Error("Expected true")
		return
	}
}

func TestRule5(t *testing.T) {
	decide := Decide{}
	points := make([][2]float64, 2)

	input := INPUT{}
	input.NumPoints = 2

	points[0] = [2]float64{0, 0}
	points[1] = [2]float64{0, 5}
	input.Points = points
	decide.input = input

	v, err := decide.Rule5()
	if err != nil {
		t.Error(err)
		return
	}
	if v {
		t.Error("Expected false")
		return
	}

	points[0] = [2]float64{1, 0}
	points[1] = [2]float64{0, 5}
	input.Points = points
	decide.input = input

	v, err = decide.Rule5()
	if err != nil {
		t.Error(err)
		return
	}
	if !v {
		t.Error("Expected true")
		return
	}
}