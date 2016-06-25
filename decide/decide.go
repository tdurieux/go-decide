package decide

import (
	"errors"
	"math"
	"reflect"
	"fmt"
)

const NB_LIC = 15

type Command string

const (
	ANDD Command = "ANDD"
	ORR Command = "NOTUSED"
	NOTUSED Command = "ORR"
)

type Parameters struct {
	RADIUS1 float64
	RADIUS2 float64
	LENGTH1 float64
	LENGTH2 float64
	DIST float64
	EPSILON float64
	QUADS int
	AREA1 float64
	AREA2 float64
	A_PTS int
	B_PTS int
	C_PTS int
	D_PTS int
	E_PTS int
	F_PTS int
	G_PTS int
	K_PTS int
	N_PTS int
	Q_PTS int
}

type INPUT struct {
	NumPoints int `json:"NUMPOINTS"`
	Points    [][2]float64 `json:"POINTS"`
	LCM       map[string][NB_LIC]Command `json:"LCM"`
	PUV       [NB_LIC]bool `json:"PUV"`
	Parameters Parameters `json:"PARAMETERS"`
}

type Pum [NB_LIC][NB_LIC]bool
type Fuv [NB_LIC]bool
type Cmv [NB_LIC]bool

type Decide struct {
	input INPUT
	Launch string `json:"LAUNCH"`
	CMV    Cmv `json:"CMV"`
	PUM    Pum `json:"PUM"`
	FUV    Fuv `json:"FUV"`
}

func (d *Decide) Decide(input INPUT) error {
	if input.NumPoints < 2 || input.NumPoints > 100 {
		return errors.New("Invalid NumPoints value.")
	}
	d.input = input

	err := d.performCMV()
	if err != nil {
		return err
	}

	err = d.performPUM()
	if err != nil {
		return err
	}

	err = d.performFUV()
	if err != nil {
		return err
	}
	d.isToLaunch()

	return nil
}

func (d *Decide) performPUM() error {
	for i := 0; i < NB_LIC; i++ {
		cmvi := d.CMV[i]
		for j := 0; j < NB_LIC; j++ {
			lcm := d.input.LCM[fmt.Sprintf("%d", i)][j]
			if (lcm == NOTUSED) {
				d.PUM[i][j] = true
				continue
			}
			cmvj := d.CMV[j]
			if (lcm == ANDD) {
				d.PUM[i][j] = cmvi && cmvj
				continue
			}
			if (lcm == ORR) {
				d.PUM[i][j] = cmvi || cmvj
				continue
			}

		}
	}
	return nil
}

func (d *Decide) performFUV() error {
	for i := 0; i < NB_LIC; i++ {
		if !d.input.PUV[i] {
			d.FUV[i] = true
			continue
		}
		fuv := true
		for j := 0; j < NB_LIC && fuv; j++ {
			if i == j {
				continue
			}
			fuv = d.PUM[i][j]
		}
		d.FUV[i] = fuv
	}
	return nil
}

func (d *Decide) performCMV() error {
	var cmv   Cmv

	for i := 0; i < NB_LIC; i++ {
		decideValue := reflect.ValueOf(d)
		methodName := fmt.Sprintf("Rule%d", i)
		method := decideValue.MethodByName(methodName)
		ruleMethod := method.Interface().(func() (bool, error))
		value, err := ruleMethod()
		if err != nil {
			return err
		}
		cmv[i] = value
	}

	d.CMV = cmv
	return nil
}

func (d Decide) Rule0() (bool, error)  {
	if d.input.Parameters.LENGTH1 < 0 {
		return false, errors.New("Invalid length1")
	}
	for i, c := range d.input.Points {
		if (i >= d.input.NumPoints - 1) {
			break;
		}
		next := d.input.Points[i + 1]
		if computeDistancePointToPoint(c, next) > d.input.Parameters.LENGTH1 {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule1() (bool, error)  {
	if d.input.Parameters.RADIUS1 < 0 {
		return false, errors.New("Invalid RADIUS1")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - 2) {
			break;
		}
		p2 := d.input.Points[i + 1]
		p3 := d.input.Points[i + 2]

		// center of the 3 points
		var pc [2]float64
		pc[0] = (p1[0] + p2[0] + p3[0]) / 3;
		pc[1] = (p1[1] + p2[1] + p3[1]) / 3;

		r1 := computeDistancePointToPoint(p1, pc)
		r2 := computeDistancePointToPoint(p2, pc)
		r3 := computeDistancePointToPoint(p3, pc)
		if (r1 > d.input.Parameters.RADIUS1 || r2 > d.input.Parameters.RADIUS1 || r3 > d.input.Parameters.RADIUS1) {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule2() (bool, error)  {
	if d.input.Parameters.EPSILON < 0 || d.input.Parameters.EPSILON >= math.Pi {
		return false, errors.New("Invalid EPSILON")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - 2) {
			break;
		}
		p2 := d.input.Points[i + 1]
		p3 := d.input.Points[i + 2]

		dp1p2 := computeDistancePointToPoint(p1, p2)
		dp2p3 := computeDistancePointToPoint(p2, p3)
		dp1p3 := computeDistancePointToPoint(p1, p3)

		// If either the first point or the last point (or both)
		// coincides with the vertex, the angle is undefined and
		// the LIC is not satisfied by those three points
		if dp1p2 == 0 || dp2p3 == 0 {
			return false, nil
		}

		angle := math.Acos((math.Pow(dp1p2, 2) + math.Pow(dp2p3, 2) - math.Pow(dp1p3, 2)) / (2 * dp1p2 * dp2p3));
		if (angle < math.Pi - d.input.Parameters.EPSILON || angle > math.Pi + d.input.Parameters.EPSILON) {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule3() (bool, error)  {
	if d.input.Parameters.AREA1 < 0 {
		return false, errors.New("Invalid AREA1")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - 2) {
			break;
		}
		p2 := d.input.Points[i + 1]
		p3 := d.input.Points[i + 2]

		area := math.Abs(((p1[0] - p3[0]) * (p2[1] - p1[1]) - (p1[0] - p2[0]) * (p3[1] - p1[1]))/2)
		area = math.Abs(p1[0] * (p2[1] - p3[1]) + p2[0] * (p3[1] - p1[1]) + p3[0] * (p1[1] - p2[1]))/2
		if area > d.input.Parameters.AREA1 {
			return true, nil
		}

	}
	return false, nil
}

func (d Decide) Rule4() (bool, error)  {
	for i, _ := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.Q_PTS) {
			break;
		}
		lieCounter := 0
		currentQuad := 1
		for ndx := i; ndx < (i + d.input.Parameters.Q_PTS); ndx++ {
			if (getQuadranNumber(d.input.Points[ndx]) != currentQuad) {
				lieCounter++;
				if (lieCounter > d.input.Parameters.QUADS) {
					return true, nil
				}
			}

			currentQuad++;
			if (currentQuad > 4) {
				currentQuad = 1;
			}
		}

		if (lieCounter > 3) {
			break;
		}
	}
	return false, nil
}

func (d Decide) Rule5() (bool, error)  {
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - 1) {
			break;
		}
		p2 := d.input.Points[i + 1]

		if (p2[0] - p1[0] < 0) {
			return true, nil
		}

	}
	return false, nil
}

func (d Decide) Rule6() (bool, error)  {
	if d.input.NumPoints < 3 {
		return false, nil
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.N_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.N_PTS]

		dp1p2 := computeDistancePointToPoint(p1, p2)
		if dp1p2 == 0 {
			for j := i; j < i + d.input.Parameters.N_PTS; j++ {
				if (computeDistancePointToPoint(d.input.Points[j], p1) > d.input.Parameters.DIST) {
					return true, nil
				}
			}
		} else {
			for j := i; j < i + d.input.Parameters.N_PTS; j++ {
				if (computeDistancePointToLine(d.input.Points[j], computeEquationLine(p1, p2)) > d.input.Parameters.DIST) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (d Decide) Rule7() (bool, error)  {
	if d.input.NumPoints < 3 {
		return false, nil
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.K_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.K_PTS]
		if computeDistancePointToPoint(p1, p2) > d.input.Parameters.LENGTH1 {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule8() (bool, error)  {
	if d.input.NumPoints < 5 {
		return false, nil
	}
	if d.input.Parameters.A_PTS + d.input.Parameters.B_PTS > d.input.NumPoints - 3 {
		return false, nil
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.A_PTS - d.input.Parameters.B_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.A_PTS]
		p3 := d.input.Points[i + d.input.Parameters.A_PTS + d.input.Parameters.B_PTS]

		// center of the 3 points
		var pc [2]float64
		pc[0] = (p1[0] + p2[0] + p3[0]) / 3;
		pc[1] = (p1[1] + p2[1] + p3[1]) / 3;

		r1 := computeDistancePointToPoint(p1, pc)
		r2 := computeDistancePointToPoint(p2, pc)
		r3 := computeDistancePointToPoint(p3, pc)
		if (r1 > d.input.Parameters.RADIUS1 || r2 > d.input.Parameters.RADIUS1 || r3 > d.input.Parameters.RADIUS1) {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule9() (bool, error)  {
	if d.input.NumPoints < 5 {
		return false, nil
	}
	if d.input.Parameters.C_PTS + d.input.Parameters.D_PTS > d.input.NumPoints - 3 {
		return false, nil
	}
	if d.input.Parameters.EPSILON < 0 || d.input.Parameters.EPSILON >= math.Pi {
		return false, errors.New("Invalid EPSILON")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.C_PTS - d.input.Parameters.D_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.C_PTS]
		p3 := d.input.Points[i + d.input.Parameters.C_PTS + d.input.Parameters.D_PTS]

		dp1p2 := computeDistancePointToPoint(p1, p2)
		dp2p3 := computeDistancePointToPoint(p2, p3)
		dp1p3 := computeDistancePointToPoint(p1, p3)

		// If either the first point or the last point (or both)
		// coincides with the vertex, the angle is undefined and
		// the LIC is not satisfied by those three points
		if dp1p2 == 0 || dp2p3 == 0 {
			continue
		}

		angle := math.Acos((math.Pow(dp1p2, 2) + math.Pow(dp2p3, 2) - math.Pow(dp1p3, 2)) / (2 * dp1p2 * dp2p3));
		if (angle < math.Pi - d.input.Parameters.EPSILON || angle > math.Pi + d.input.Parameters.EPSILON) {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule10() (bool, error)  {
	if d.input.Parameters.AREA1 < 0 {
		return false, errors.New("Invalid AREA1")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.E_PTS - d.input.Parameters.F_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.E_PTS]
		p3 := d.input.Points[i + d.input.Parameters.E_PTS + d.input.Parameters.F_PTS]

		area := math.Abs((p1[0] * (p2[1] - p3[1]) + p2[0] * (p3[1] - p2[1]) + p3[0] * (p1[1] - p2[1]))/2)
		if area > d.input.Parameters.AREA1 {
			return true, nil
		}

	}
	return false, nil
}

func (d Decide) Rule11() (bool, error)  {
	if d.input.NumPoints < 3 {
		return false, nil
	}
	if d.input.Parameters.LENGTH2 < 0 {
		return false, errors.New("Invalid LENGTH2")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.K_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.K_PTS]

		dp1p2 := computeDistancePointToPoint(p1, p2)
		if (dp1p2 > d.input.Parameters.LENGTH1) {
			for i, p1 := range d.input.Points {
				if (i >= d.input.NumPoints - d.input.Parameters.K_PTS) {
					break;
				}
				p2 := d.input.Points[i + d.input.Parameters.K_PTS]

				dp1p2 := computeDistancePointToPoint(p1, p2)
				if (dp1p2 < d.input.Parameters.LENGTH2) {
					return true, nil
				}
			}
		}

	}
	return false, nil
}

func (d Decide) Rule12() (bool, error)  {
	cond1 := false
	cond2 := false
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.K_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.K_PTS]
		dp1dp2 := computeDistancePointToPoint(p1, p2)
		if !cond1 && dp1dp2 > d.input.Parameters.LENGTH1 {
			cond1 = true
		}
		if !cond2 && dp1dp2 > d.input.Parameters.LENGTH2 {
			cond2 = true
		}
		if cond1 && cond2 {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule13() (bool, error)  {
	cond1 := false
	cond2 := false
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.A_PTS - d.input.Parameters.B_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.A_PTS]
		p3 := d.input.Points[i + d.input.Parameters.A_PTS + d.input.Parameters.B_PTS]

		// center of the 3 points
		var pc [2]float64
		pc[0] = (p1[0] + p2[0] + p3[0]) / 3;
		pc[1] = (p1[1] + p2[1] + p3[1]) / 3;

		r1 := computeDistancePointToPoint(p1, pc)
		r2 := computeDistancePointToPoint(p2, pc)
		r3 := computeDistancePointToPoint(p3, pc)
		if (!cond2 && (r1 > d.input.Parameters.RADIUS2 || r2 > d.input.Parameters.RADIUS2 || r3 > d.input.Parameters.RADIUS2)) {
			cond2 = true
		}

		if (!cond1 && (r1 > d.input.Parameters.RADIUS1 || r2 > d.input.Parameters.RADIUS1 || r3 > d.input.Parameters.RADIUS1)) {
			cond1 = true
		}
		if cond1 && cond2 {
			return true, nil
		}

	}
	return false, nil
}

func (d Decide) Rule14() (bool, error)  {
	cond1 := false
	cond2 := false
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.E_PTS - d.input.Parameters.F_PTS) {
			break;
		}
		p2 := d.input.Points[i + d.input.Parameters.E_PTS]
		p3 := d.input.Points[i + d.input.Parameters.E_PTS + d.input.Parameters.F_PTS]

		area := math.Abs((p1[0] * (p2[1] - p3[1]) + p2[0] * (p3[1] - p2[1]) + p3[0] * (p1[1] - p2[1]))/2)
		if !cond1 && area > d.input.Parameters.AREA1 {
			cond1 = true
		}
		if !cond2 && area > d.input.Parameters.AREA2{
			cond2 = true
		}
		if cond1 && cond2 {
			return true, nil
		}
	}
	return false, nil
}


func (d *Decide) isToLaunch()  {
	for _, v := range d.FUV {
		if (!v) {
			d.Launch = "NO"
			return
		}
	}
	d.Launch = "YES"
}

func computeEquationLine(p1 [2]float64, p2 [2]float64) [3]float64  {
	var equation [3]float64

	equation[0] = p1[1] - p2[1];
	equation[1] = p1[0] - p2[0];
	equation[2] = p1[0] * p2[1] - p2[0] * p1[1];

	return equation
}

func computeDistancePointToLine(p1 [2]float64, equationLine [3]float64) float64  {
	return math.Abs(equationLine[0] * p1[0] + equationLine[1] * p1[1] + equationLine[2]) / math.Sqrt(math.Pow(equationLine[0], 2) + math.Pow(equationLine[1], 2));
}

func computeDistancePointToPoint(p1 [2]float64, p2 [2]float64) float64  {
	return math.Sqrt(math.Pow(p1[0] - p2[0], 2) + math.Pow(p1[1] - p2[1], 2))
}

func getQuadranNumber(p [2]float64) int {
	x := p[0]
	y := p[1]

	if (x >= 0 && y >= 0) {
		return 1;
	}
	if (x < 0 && y >= 0) {
		return 2;
	}
	if (x < 0 && y < 0) {
		return 3;
	}
	return 4;
}