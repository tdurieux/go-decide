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
	if (len(input.Points) != input.NumPoints) {
		return errors.New("Invalid NumPoints value different from the actual number of points.")
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

// There exists at least one set of two consecutive data points
// that are a distance greater than the length, LENGTH1, apart.
func (d Decide) Rule0() (bool, error)  {
	// (0 ≤ LENGTH1)
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

// There exists at least one set of three consecutive data points
// that cannot all be contained within or on a circle of radius RADIUS1.
func (d Decide) Rule1() (bool, error)  {
	// (0 ≤ RADIUS1)
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

// There exists at least one set of three consecutive data points which
// form an angle such that:  angle < (PI−EPSILON)  or angle > (PI+EPSILON)
// The second of the three consecutive points is always the vertex of the angle.
// If either the first point or the last point (or both) coincides with the vertex,
// the angle is undefined and the LIC is not satisfied by those three points
func (d Decide) Rule2() (bool, error)  {
	// (0 ≤ EPSILON < PI)
	if d.input.Parameters.EPSILON < 0 || d.input.Parameters.EPSILON >= math.Pi {
		return false, errors.New("Invalid EPSILON")
	}
	for i, a := range d.input.Points {
		if (i >= d.input.NumPoints - 2) {
			break;
		}
		b := d.input.Points[i + 1]
		c := d.input.Points[i + 2]

		// http://stackoverflow.com/questions/3486172/angle-between-3-points
		ab := [2]float64{b[0] - a[0], b[1] - a[1]}
		cb := [2]float64{b[0] - c[0], b[1] - c[1]}

		dot := (ab[0] * cb[0] + ab[1] * cb[1])
		cross := (ab[0] * cb[1] - ab[1] * cb[0])

		angle := math.Atan2(cross, dot)

		// If either the first point or the last point (or both)
		// coincides with the vertex, the angle is undefined and
		// the LIC is not satisfied by those three points
		if a == b || b == c {
			return false, nil
		}
		if (angle < math.Pi - d.input.Parameters.EPSILON || angle > math.Pi + d.input.Parameters.EPSILON) {
			return true, nil
		}
	}
	return false, nil
}

// There exists at least one set of three consecutive data points
// that are the vertices of a triangle with area greater than AREA1
func (d Decide) Rule3() (bool, error)  {
	// (0 ≤ AREA1)
	if d.input.Parameters.AREA1 < 0 {
		return false, errors.New("Invalid AREA1")
	}
	for i, p1 := range d.input.Points {
		if (i >= d.input.NumPoints - 2) {
			break;
		}
		p2 := d.input.Points[i + 1]
		p3 := d.input.Points[i + 2]

		area := math.Abs(p1[0] * (p2[1] - p3[1]) + p2[0] * (p3[1] - p1[1]) + p3[0] * (p1[1] - p2[1]))/2
		if area > d.input.Parameters.AREA1 {
			return true, nil
		}

	}
	return false, nil
}

// There exists at least one set of Q PTS consecutive data points
// that lie in more than QUADS quadrants.
// Where there is ambiguity as to which quadrant contains a given point, priority
// of decision will be by quadrant number, i.e., I, II, III, IV.
// For example, the data point (0,0) is in quadrant I, the point (-l,0) is in quadrant II,
// the point (0,-l) is in quadrant III, the point  (0,1) is in quadrant I and the point (1,0) is in quadrant I.
func (d Decide) Rule4() (bool, error)  {
	// (2 ≤ Q PTS ≤ NUMPOINTS)
	if d.input.Parameters.Q_PTS < 2 || d.input.Parameters.Q_PTS > d.input.NumPoints {
		return false, errors.New("Invalid Q_PTS")
	}
	// (1 ≤ QUADS ≤ 3)
	if d.input.Parameters.QUADS < 1 || d.input.Parameters.QUADS > 3 {
		return false, errors.New("Invalid QUADS")
	}
	for i, _ := range d.input.Points {
		if (i > d.input.NumPoints - d.input.Parameters.Q_PTS) {
			break;
		}
		usedQuadrants := make([]bool, 4)
		for ndx := i; ndx < (i + d.input.Parameters.Q_PTS); ndx++ {
			usedQuadrants[getQuadranNumber(d.input.Points[ndx])] = true
		}
		countUsed := 0
		for _, v := range usedQuadrants {
			if (v) {
				countUsed++;
				// lie in more than QUADS quadrants
				if (countUsed > d.input.Parameters.QUADS) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// There exists at least one set of two consecutive data points,
// (X[i],Y[i]) and (X[j],Y[j]), such that X[j] - X[i] < 0. (where i = j-1)
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

// There exists at least one set of N PTS consecutive data points such
// that at least one of the points lies a distance greater than DIST
// from the line joining the first and last of these N PTS points.
// If the first and last points of these N PTS are identical,
// then the calculated distance to compare with DIST will be the distance
// from the coincident point to all other points of the N PTS consecutive points.
// The condition is not met when NUMPOINTS < 3.
func (d Decide) Rule6() (bool, error)  {
	// The condition is not met when NUMPOINTS < 3.
	if d.input.NumPoints < 3 {
		return false, nil
	}
	// (3 ≤ N PTS ≤ NUMPOINTS)
	if d.input.Parameters.N_PTS < 3 || d.input.Parameters.N_PTS > d.input.NumPoints {
		return false, errors.New("Invalid N_PTS.")
	}
	// (0 ≤ DIST)
	if d.input.Parameters.DIST < 0 {
		return false, errors.New("Invalid DIST.")
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

// There exists at least one set of two data points separated by exactly K PTS consecutive intervening
// points that are a distance greater than the length, LENGTH1, apart.
// The condition is not met when NUMPOINTS < 3.
func (d Decide) Rule7() (bool, error)  {
	// The condition is not met when NUMPOINTS < 3.
	if d.input.NumPoints < 3 {
		return false, nil
	}
	// 1 ≤ K PTS ≤ (NUMPOINTS−2)
	if d.input.Parameters.K_PTS < 1 || d.input.Parameters.K_PTS > d.input.NumPoints - 2 {
		return false, errors.New("Invalid K_PTS.")
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

// There exists at least one set of three data points separated by exactly A PTS and B PTS
// consecutive intervening points, respectively, that cannot be contained within or on a circle of
// radius RADIUS1. The condition is not met when NUMPOINTS < 5.
func (d Decide) Rule8() (bool, error)  {
	// The condition is not met when NUMPOINTS < 5.
	if d.input.NumPoints < 5 {
		return false, nil
	}
	// A PTS+B PTS ≤ (NUMPOINTS−3)
	if d.input.Parameters.A_PTS + d.input.Parameters.B_PTS > d.input.NumPoints - 3 {
		return false, errors.New("Invalid A_PTS, B_PTS.")
	}
	// 1 ≤ A PTS
	if d.input.Parameters.A_PTS < 1 {
		return false, errors.New("Invalid A_PTS.")
	}
	// 1 ≤ B PTS
	if d.input.Parameters.B_PTS < 1 {
		return false, errors.New("Invalid B_PTS.")
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
	if d.input.Parameters.C_PTS < 1 {
		return false, errors.New("Invalid C_PTS.")
	}
	if d.input.Parameters.D_PTS < 1 {
		return false, errors.New("Invalid D_PTS.")
	}
	if d.input.Parameters.C_PTS + d.input.Parameters.D_PTS > d.input.NumPoints - 3 {
		return false, nil
	}
	if d.input.Parameters.EPSILON < 0 || d.input.Parameters.EPSILON >= math.Pi {
		return false, errors.New("Invalid EPSILON")
	}
	for i, a := range d.input.Points {
		if (i >= d.input.NumPoints - d.input.Parameters.C_PTS - d.input.Parameters.D_PTS) {
			break;
		}
		b := d.input.Points[i + d.input.Parameters.C_PTS]
		c := d.input.Points[i + d.input.Parameters.C_PTS + d.input.Parameters.D_PTS]

		// http://stackoverflow.com/questions/3486172/angle-between-3-points
		ab := [2]float64{b[0] - a[0], b[1] - a[1]}
		cb := [2]float64{b[0] - c[0], b[1] - c[1]}

		dot := (ab[0] * cb[0] + ab[1] * cb[1])
		cross := (ab[0] * cb[1] - ab[1] * cb[0])

		angle := math.Atan2(cross, dot)

		// If either the first point or the last point (or both)
		// coincides with the vertex, the angle is undefined and
		// the LIC is not satisfied by those three points
		if a == b || b == c {
			return false, nil
		}
		if (angle < math.Pi - d.input.Parameters.EPSILON || angle > math.Pi + d.input.Parameters.EPSILON) {
			return true, nil
		}
	}
	return false, nil
}

func (d Decide) Rule10() (bool, error)  {
	if d.input.Parameters.E_PTS < 1 {
		return false, errors.New("Invalid E_PTS.")
	}
	if d.input.Parameters.F_PTS < 1 {
		return false, errors.New("Invalid F_PTS.")
	}
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
	if d.input.Parameters.K_PTS < 1 {
		return false, errors.New("Invalid K_PTS.")
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
	if d.input.Parameters.K_PTS < 1 {
		return false, errors.New("Invalid K_PTS.")
	}
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
	if d.input.Parameters.A_PTS < 1 {
		return false, errors.New("Invalid A_PTS.")
	}
	if d.input.Parameters.B_PTS < 1 {
		return false, errors.New("Invalid B_PTS.")
	}
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
	if d.input.Parameters.E_PTS < 1 {
		return false, errors.New("Invalid E_PTS.")
	}
	if d.input.Parameters.F_PTS < 1 {
		return false, errors.New("Invalid F_PTS.")
	}
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
		return 0;
	}
	if (x < 0 && y >= 0) {
		return 1;
	}
	if (x < 0 && y < 0) {
		return 2;
	}
	return 3;
}