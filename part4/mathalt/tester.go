package mathalt

import (
	"fmt"
	"math"
	"slices"
)

// Reference tables printed via numpy. See reftable.py script next to these
// files.
var (
	refSin = []referenceEntry{
		{-1.5707963267948966, -1.0000000000000000},
		{-1.2982580992455555, -0.9630907674188940},
		{-1.2184943426914576, -0.9385808793143084},
		{-1.1957550763274201, -0.9304925122075935},
		{-1.1780972450961724, -0.9238795325112867},
		{-1.1409417229533199, -0.9090263514801161},
		{-0.8841377592160081, -0.7733686514744063},
		{-0.7853981633974483, -0.7071067811865475},
		{-0.6720153354048302, -0.6225643884308514},
		{-0.3926990816987241, -0.3826834323650898},
		{0.0000000000000000, 0.0000000000000000},
		{0.0116258464893433, 0.0116255845989505},
		{0.0475698891355461, 0.0475519502264479},
		{0.2419748075394046, 0.2396203672324879},
		{0.3926990816987241, 0.3826834323650898},
		{0.6528191758517843, 0.6074282980497632},
		{0.7853981633974483, 0.7071067811865475},
		{0.8760930078932114, 0.7682436582429083},
		{1.0373821250379316, 0.8610760522777925},
		{1.0958561018790283, 0.8893200574785602},
		{1.1631391187652618, 0.9180521638076899},
		{1.1780972450961724, 0.9238795325112867},
		{1.4241798848957483, 0.9892710496246157},
		{1.5707963267948966, 1.0000000000000000},
	}
	refCos = []referenceEntry{
		{-1.5707963267948966, 0.0000000000000001},
		{-1.2982580992455555, 0.2691768446811238},
		{-1.2184943426914576, 0.3450593180680382},
		{-1.1957550763274201, 0.3663109126488062},
		{-1.1780972450961724, 0.3826834323650898},
		{-1.1409417229533199, 0.4167386378952020},
		{-0.8841377592160081, 0.6339565670585473},
		{-0.7853981633974483, 0.7071067811865476},
		{-0.6720153354048302, 0.7825685799070392},
		{-0.3926990816987241, 0.9238795325112867},
		{0.0000000000000000, 1.0000000000000000},
		{0.0116258464893433, 0.9999324206078792},
		{0.0475698891355461, 0.9988687661698414},
		{0.2419748075394046, 0.9708666641755538},
		{0.3926990816987241, 0.9238795325112867},
		{0.6528191758517843, 0.7943745103717567},
		{0.7853981633974483, 0.7071067811865476},
		{0.8760930078932114, 0.6401575443354187},
		{1.0373821250379316, 0.5084761864568412},
		{1.0958561018790283, 0.4572852888146854},
		{1.1631391187652618, 0.3964596127325183},
		{1.1780972450961724, 0.3826834323650898},
		{1.4241798848957483, 0.1460917190487235},
		{1.5707963267948966, 0.0000000000000001},
	}
	refAsin = []referenceEntry{
		{0.1342060629826266, 0.1346122339292709},
		{0.2680144535895596, 0.2713314925000239},
		{0.3058372107164309, 0.3108176508297974},
		{0.3303461553080264, 0.3366702960208309},
		{0.3348104335984347, 0.3414040456651079},
		{0.3350956666354683, 0.3417067653616962},
		{0.4001219886990834, 0.4116499505167356},
		{0.4316766805471766, 0.4463507410771893},
		{0.4803885043622710, 0.5010976229400018},
		{0.4959981067744247, 0.5189839188312722},
		{0.5868987539091778, 0.6272231940743844},
		{0.6559734337670925, 0.7154715841244675},
		{0.7263847291470424, 0.8130470269416959},
		{0.7570863924387091, 0.8588417858343772},
		{0.8018690431757729, 0.9304167911715419},
		{0.8257366578545743, 0.9715069558959933},
		{0.8464659083616703, 1.0093123461637161},
		{0.8858221124120441, 1.0882627245667755},
		{0.9027719244741774, 1.1261710962578522},
		{0.9805847197831580, 1.3734213889134166},
	}
	refSqrt = []referenceEntry{
		{0.1054967087809499, 0.3248025689260322},
		{0.1505365491640086, 0.3879903982884224},
		{0.2424384301673397, 0.4923803714277608},
		{0.2537132174866612, 0.5036995309573568},
		{0.4077625011130428, 0.6385628403791148},
		{0.4732401084220416, 0.6879244932563759},
		{0.4741777523286932, 0.6886056580719426},
		{0.5403318080611316, 0.7350726549540063},
		{0.5994709426873445, 0.7742550888998693},
		{0.6394285385150210, 0.7996427568077016},
		{0.6475495472299926, 0.8047046335333186},
		{0.6587028313990854, 0.8116050957202556},
		{0.7404919920717400, 0.8605184437719741},
		{0.7746343818552135, 0.8801331614336625},
		{0.7779171807577306, 0.8819961342079288},
		{0.7782579279961580, 0.8821892812748056},
		{0.9099387223688884, 0.9539070826704708},
		{0.9113374310740314, 0.9546399483962692},
		{0.9365264686140918, 0.9677429765253230},
		{0.9370402735727158, 0.9680084057345348},
	}
)

type TestFnResult struct {
	MaxDiff   float64
	TotalDiff float64
	DiffCount int

	MaxDiffInputVal  float64
	MaxDiffOutputVal float64
	MaxDiffExpect    float64

	Label string
}

type PrecisionTester struct {
	Results []TestFnResult

	ResultCount         int
	ProgressResultCount int

	Testing      bool
	StepIndex    int
	ResultOffset int

	InputVal float64
}

type referenceEntry struct {
	input, output float64
}

func PrintSep() {
	// Marks out precision decimals in output.
	fmt.Println("   ________________             ________________")
}

func (r *TestFnResult) AvgDiff() float64 {
	if r.DiffCount == 0 {
		return 0
	}
	return r.TotalDiff / float64(r.DiffCount)
}

func (r *TestFnResult) PrintResult() {
	fmt.Printf(
		"%+.24f (%+.24f) at %+.24f [%s]\n",
		r.MaxDiff, r.AvgDiff(), r.MaxDiffInputVal, r.Label,
	)
}

// Step either starts the testing, advances it to the next step, or stops the
// testing.
func (t *PrecisionTester) Step(minInput, maxInput float64, stepCount int) bool {
	if t.Testing {
		// Step
		t.StepIndex++
	} else {
		// Start
		t.Testing = true
		t.StepIndex = 0
	}

	if t.StepIndex < stepCount {
		// Advance input value
		t.ResultOffset = 0

		tStep := float64(t.StepIndex) / float64(stepCount-1)
		t.InputVal = (1-tStep)*minInput + tStep*maxInput
	} else {
		// Stop
		t.ResultCount += t.ResultOffset

		if t.ProgressResultCount < t.ResultCount {
			fmt.Println("   Largest                      Average")
			PrintSep()
			for t.ProgressResultCount < t.ResultCount {
				t.Results[t.ProgressResultCount].PrintResult()
				t.ProgressResultCount++
			}
		}

		t.Testing = false
	}
	return t.Testing
}

func (t *PrecisionTester) Test(expected, output float64, format string, args ...any) {
	resultIndex := t.ResultCount + t.ResultOffset

	var res *TestFnResult
	if t.StepIndex == 0 {
		t.Results = append(t.Results, TestFnResult{})
		res = &t.Results[len(t.Results)-1]
		res.Label = fmt.Sprintf(format, args...)
	} else {
		res = &t.Results[resultIndex]
	}

	diff := math.Abs(expected - output)
	res.TotalDiff += diff
	res.DiffCount++

	if res.MaxDiff < diff {
		res.MaxDiff = diff
		res.MaxDiffInputVal = t.InputVal
		res.MaxDiffOutputVal = output
		res.MaxDiffExpect = expected
	}

	t.ResultOffset++
}

func (t *PrecisionTester) PrintResults() {
	fmt.Printf("\nSorted results by maximum error:\n")
	fmt.Println("   Largest                      Average")
	PrintSep()

	ranking := make([]int, len(t.Results))
	for i := range ranking {
		ranking[i] = i
	}
	slices.SortFunc(ranking, func(a, b int) int {
		ra := t.Results[a]
		rb := t.Results[b]
		switch {
		case ra.MaxDiff < rb.MaxDiff:
			return -1
		case ra.MaxDiff > rb.MaxDiff:
			return 1
		case ra.TotalDiff < rb.TotalDiff:
			return -1
		case ra.TotalDiff > rb.TotalDiff:
			return 1
		default:
			return 0
		}
	})

	for ri := 0; ri < len(ranking); ri++ {
		res := t.Results[ranking[ri]]

		fmt.Printf("%+.24f (%+.24f) [%s", res.MaxDiff, res.AvgDiff(), res.Label)

		// Continue to append equal results, if any.
		for rj := ri + 1; rj < len(ranking); rj++ {
			nextRes := t.Results[ranking[rj]]
			if nextRes.MaxDiff == res.MaxDiff && nextRes.TotalDiff == res.TotalDiff {
				fmt.Printf(", %s", nextRes.Label)
				ri++
			}
			break
		}
		fmt.Printf("]\n")
	}
}

func TestFunctions() {
	funcs := []struct {
		name    string
		fn      func(float64) float64
		ref     []referenceEntry
		refImpl func(float64) float64
		x0, x1  float64
	}{
		{"SinQ", SinQ, refSin, math.Sin, -pi, pi},
		{"SinAlt", SinAlt, refSin, math.Sin, -pi, pi},
		{"SinTaylor3", SinTaylorN(3), refSin, math.Sin, -pi, pi},
		{"SinTaylor4", SinTaylorN(4), refSin, math.Sin, -pi, pi},
		{"SinTaylor5", SinTaylorN(5), refSin, math.Sin, -pi, pi},
		{"SinTaylor6", SinTaylorN(6), refSin, math.Sin, -pi, pi},
		{"SinTaylor7", SinTaylorN(7), refSin, math.Sin, -pi, pi},
		{"SinTaylor8", SinTaylorN(8), refSin, math.Sin, -pi, pi},
		{"SinTaylor9", SinTaylorN(9), refSin, math.Sin, -pi, pi},
		{"SinTaylor10", SinTaylorN(10), refSin, math.Sin, -pi, pi},
		{"SinTaylor11", SinTaylorN(11), refSin, math.Sin, -pi, pi},
		{"CosAlt", CosAlt, refCos, math.Cos, -pi / 2, pi / 2},
		{"AsinAlt", AsinAlt, refAsin, math.Asin, 0, 1},
		{"SqrtAlt", SqrtAlt, refSqrt, math.Sqrt, 0, 1},
	}

	for i, tt := range funcs {
		if i != 0 {
			fmt.Println()
		}
		TestReferenceTable(tt.name, tt.fn, tt.ref)
	}

	pt := PrecisionTester{}
	for _, tt := range funcs {
		for pt.Step(tt.x0, tt.x1, 10_000_000) {
			pt.Test(tt.refImpl(pt.InputVal), tt.fn(pt.InputVal), "%s", tt.name)
		}
	}

	pt.PrintResults()
}

func TestSinFunctions() {
	type fnDef struct {
		name string
		fn   func(float64) float64
	}
	funcs := []fnDef{
		{"SinQ", SinQ},
		{"SinAlt", SinAlt},
	}

	for n := 3; n < 20; n++ {
		funcs = append(
			funcs,
			//fnDef{fmt.Sprintf("SinTaylor%d", n), SinTaylorN(n)},
			//fnDef{fmt.Sprintf("SinTaylorHorner%d", n), SinTaylorHornerN(n)},
			//fnDef{fmt.Sprintf("SinTaylorHornerFMA%d", n), SinTaylorHornerFMAN(n)},
			//fnDef{fmt.Sprintf("SinTaylorHornerFMAAlt%d", n), SinTaylorHornerFMAAltN(n)},
			fnDef{fmt.Sprintf("SinTaylorPre%d", n), SinTaylorFunc(SinTaylorPre, n)},
		)
	}

	for n := 3; n < 12; n++ {
		funcs = append(
			funcs,
			fnDef{fmt.Sprintf("SinMFTWP%d", n), SinTaylorFunc(SinMFTWP, n)},
		)
	}
	funcs = append(funcs, fnDef{"SinMFTWP_Manual9", SinMFTWP_Manual9})

	for i, tt := range funcs {
		if i != 0 {
			fmt.Println()
		}
		TestReferenceTable(tt.name, tt.fn, refSin)
	}

	var pt PrecisionTester
	for pt.Step(-pi, pi, 10_000_000) {
		for _, tt := range funcs {
			pt.Test(math.Sin(pt.InputVal), tt.fn(pt.InputVal), "%s", tt.name)
		}
	}

	pt.PrintResults()
}

func TestReferenceTable(name string, fn func(float64) float64, reference []referenceEntry) {
	fmt.Printf("\n=== %s ===\n", name)
	for _, ref := range reference {
		fmt.Printf("func(%+.24f) = %+.24f [reference]\n", ref.input, ref.output)
		got := fn(ref.input)
		fmt.Printf("                                  = %+.24f [%+.24f] [%s]\n", got, ref.output-got, name)
	}
}
