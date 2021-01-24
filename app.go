package main

import (
	"flag"
	"fmt"
	"math"
	"os"
)

// R is an ideal gas constant
const R = 0.0831

// Temperature represents gas temperature
type Temperature float64

// GasVolume is the amount of gas in liters
type GasVolume float64

// CylinderVolume is cylinder size in liters
type CylinderVolume float64

// GasWeight is the amount of of gas in kilograms (kg)
type GasWeight float64

// PressureBar represents pressure (in bar)
type PressureBar float64

// AtomicWeight is an atomic weight for an element
type AtomicWeight float64

// PressureFromVolumes returns a new PressureBar instance from gas volume and cylinder volume.
func PressureFromVolumes(gasVolume GasVolume, totalVolume CylinderVolume) PressureBar {
	return PressureBar(float64(gasVolume) / float64(totalVolume))
}

// PartialPressure returns a new partial pressure object from pressure and multiplier.
func (p PressureBar) PartialPressure(pp float64) PressureBar {
	return PressureBar(float64(p) * pp)
}

// GasWeightFromMole calculates gas weight based on the mole count and atomic weight.
func GasWeightFromMole(moleCount MoleCount, atomicWeight AtomicWeight) GasWeight {
	return GasWeight(float64(moleCount) * float64(atomicWeight))
}

// MoleCount represents number of atoms
type MoleCount float64

// GasSystem is the system used to calculate amount of the gas.
type GasSystem int

const (
	// IdealGas uses ideal gas equations which do not compensate for pressure and temperature
	IdealGas GasSystem = iota
	// VanDerWaals uses Van Der Waals equations to compensate for temperature and pressure.
	VanDerWaals
)

// Gas represents various gases cylinders may contain.
type Gas int

// Available gases
const (
	Helium Gas = iota
	Oxygen
	Nitrogen
	Argon
	Neon
	Hydrogen
)

// VanDerWaalsConstant represents Van der Waals equation constants
type VanDerWaalsConstant struct {
	A float64
	B float64
}

// AtomicWeightLookup is the weight of a single mole in grams.
var AtomicWeightLookup = map[Gas]AtomicWeight{
	Argon:    14.0067,
	Helium:   4.002602,
	Hydrogen: 1.00784,
	Neon:     20.1797,
	Nitrogen: 14.0067,
	Oxygen:   15.999,
}

// VanDerWaalsConstants holds Van der Waals constants for gases.
var VanDerWaalsConstants = map[Gas]VanDerWaalsConstant{
	Argon:    {A: 1.355, B: 0.03201},
	Helium:   {A: 0.0346, B: 0.0238},
	Hydrogen: {A: 0.2476, B: 0.02661},
	Neon:     {A: 0.2135, B: 0.01709},
	Nitrogen: {A: 1.370, B: 0.0387},
	Oxygen:   {A: 1.382, B: 0.03186},
}

// GasComposition stores information about gases currently being processed
type GasComposition map[Gas]float64

// Equalize equalizes all input cylinders
func Equalize(cylinders []*Cylinder, gasSystem GasSystem, gasComposition GasComposition, temperature Temperature, verbose bool, debug bool) {
	var totalVolume CylinderVolume
	var pressureAfterEqualize PressureBar
	if gasSystem == IdealGas {
		var totalGasVolume GasVolume
		for i := range cylinders {
			totalGasVolume += cylinders[i].GasVolume(gasSystem, gasComposition, temperature)
			totalVolume += cylinders[i].CylinderVolume
		}
		pressureAfterEqualize = PressureFromVolumes(totalGasVolume, totalVolume)
	} else {
		var totalMoles MoleCount
		for i := range cylinders {
			moles := cylinders[i].Moles(temperature, gasComposition)
			if debug {
				fmt.Println("Cylinder", cylinders[i], "moles", moles)
			}
			totalMoles += moles
			totalVolume += cylinders[i].CylinderVolume
		}

		pressureAfterEqualize = cylinderMolesToPressure(totalVolume, totalMoles, temperature, gasComposition)

		if debug {
			fmt.Println("Moles:", totalMoles, "Pressure after equalize:", pressureAfterEqualize)
		}
	}

	for i := range cylinders {
		cylinders[i].Pressure = pressureAfterEqualize
	}
}

// CylinderConfiguration holds information about available cylinders and cylinder configuration, such as manifolds
type CylinderConfiguration struct {
	DestinationCylinderIsTwinset bool
	DestinationCylinderPressure  PressureBar
	DestinationCylinderVolume    CylinderVolume
	SourceCylinderIsTwinset      bool
	SourceCylinderPressure       PressureBar
	SourceCylinderVolume         CylinderVolume
}

// Cylinder represents a single cylinder and gas it contains
type Cylinder struct {
	Description    string
	CylinderVolume CylinderVolume
	Pressure       PressureBar
}

// GasVolume returns amount of gas in the cylinder
func (c1 Cylinder) GasVolume(gasSystem GasSystem, gasComposition GasComposition, temperature Temperature) GasVolume {
	if gasSystem == IdealGas {
		return GasVolume(float64(c1.CylinderVolume) * float64(c1.Pressure))
	}
	return GasVolume(gasCompositionToMoles(c1.CylinderVolume, c1.Pressure, temperature, gasComposition) * 22.4)
}

// Equalize equalizes two cylinders
func (c1 *Cylinder) Equalize(c2 *Cylinder, gasSystem GasSystem, gasComposition GasComposition, temperature Temperature, verbose bool, debug bool) {
	listOfCylinders := []*Cylinder{c1, c2}
	Equalize(listOfCylinders, gasSystem, gasComposition, temperature, verbose, debug)
}

// Moles returns number of atoms (in mole) inside a cylinder
func (c1 *Cylinder) Moles(temperature Temperature, gasComposition GasComposition) MoleCount {
	return gasCompositionToMoles(c1.CylinderVolume, c1.Pressure, temperature, gasComposition)
}

func gasCompositionToMoles(cylinderVolume CylinderVolume, cylinderPressure PressureBar, temperature Temperature, gasComposition GasComposition) MoleCount {
	var moles MoleCount
	for gasType, gasInfo := range gasComposition {
		moles += GasToMoles(cylinderVolume, cylinderPressure.PartialPressure(gasInfo), VanDerWaalsConstants[gasType], temperature)
	}
	return moles
}

// GasToMoles calculates number of atoms in given cylinder
func GasToMoles(cylinderVolume CylinderVolume, cylinderPressure PressureBar, vdwConstants VanDerWaalsConstant, temperature Temperature) MoleCount {
	a := vdwConstants.A
	b := vdwConstants.B
	P := float64(cylinderPressure)

	a2 := math.Pow(a, 2.0)
	a3 := math.Pow(a, 3.0)
	b2 := math.Pow(b, 2.0)
	V := float64(cylinderVolume)
	V2 := math.Pow(V, 2.0)
	V3 := math.Pow(V, 3.0)
	T := float64(temperature)

	subterm1 := 2*a3*V3 + 18*a2*b2*P*V3 - 9*a2*b*R*T*V3
	subterm2 := 3*a*b*(b*P*V2+R*T*V2) - a2*V2
	subterm3 := math.Pow(
		(subterm1 +
			math.Sqrt(4*math.Pow(subterm2, 3.0)+math.Pow(subterm1, 2.0))),
		(1 / 3.0))
	term1 := 0.26457 * subterm3
	term2 := a * b * subterm3
	term3 := 0.41997 * subterm2
	return MoleCount(term1/(a*b) - term3/term2 + (0.33333*V)/b)
}

// GasWeight returns weight of the gas stored inside the cylinder
func (c1 Cylinder) GasWeight(gasComposition GasComposition, temperature Temperature) GasWeight {
	var weightSum GasWeight
	for gasType, gasInfo := range gasComposition {
		moleCount := GasToMoles(c1.CylinderVolume, c1.Pressure.PartialPressure(gasInfo), VanDerWaalsConstants[gasType], temperature)
		gasWeight := GasWeightFromMole(moleCount, AtomicWeightLookup[gasType])
		weightSum += gasWeight
	}
	return weightSum
}

func cylinderMolesToPressure(cylinderVolume CylinderVolume, n MoleCount, temperature Temperature, gasComposition GasComposition) PressureBar {
	var pressureSum PressureBar
	for gasType, gasInfo := range gasComposition {
		pressureSum += MolesToPressure(cylinderVolume, MoleCount(float64(n)*gasInfo), temperature, VanDerWaalsConstants[gasType])
	}
	return pressureSum
}

// MolesToPressure returns pressure based on the volume, atomic count and gas composition.
func MolesToPressure(cylinderVolume CylinderVolume, moleCount MoleCount, T Temperature, vdwConstants VanDerWaalsConstant) PressureBar {
	V := float64(cylinderVolume)
	a := vdwConstants.A
	b := vdwConstants.B
	n := float64(moleCount)
	V2 := math.Pow(V, 2.0)
	return PressureBar(n * (-(a*n)/V2 - (R*float64(T))/(b*n-V)))
}

// CylinderList is a list of cylinders
type CylinderList []Cylinder

// TotalVolume returns total cylinder volume for all listed cylinders
func (cl CylinderList) TotalVolume() CylinderVolume {
	var totalVolume CylinderVolume
	for _, cylinder := range cl {
		totalVolume += cylinder.CylinderVolume
	}
	return totalVolume
}

// TotalGasWeight calculates the weight of the gas for all cylinders in cylinder list.
func (cl CylinderList) TotalGasWeight(gasComposition GasComposition, temperature Temperature) GasWeight {
	var weightSum GasWeight
	for _, cylinder := range cl {
		weightSum += cylinder.GasWeight(gasComposition, temperature)
	}
	return weightSum
}

// TotalGasVolume returns total gas volume for all listed cylinders
func (cl CylinderList) TotalGasVolume(gasSystem GasSystem, gasComposition GasComposition, temperature Temperature) GasVolume {
	var totalGasVolume GasVolume
	for _, cylinder := range cl {
		totalGasVolume += cylinder.GasVolume(gasSystem, gasComposition, temperature)
	}
	return totalGasVolume
}

func initializeCylinders(cylinderConfiguration CylinderConfiguration, sourceCylinders *CylinderList, destinationCylinders *CylinderList) {
	if cylinderConfiguration.SourceCylinderIsTwinset {
		*sourceCylinders = []Cylinder{
			{
				Description:    "left",
				CylinderVolume: CylinderVolume(cylinderConfiguration.SourceCylinderVolume / 2),
				Pressure:       cylinderConfiguration.SourceCylinderPressure,
			},
			{
				Description:    "right",
				CylinderVolume: CylinderVolume(cylinderConfiguration.SourceCylinderVolume / 2),
				Pressure:       cylinderConfiguration.SourceCylinderPressure,
			},
		}
	} else {
		*sourceCylinders = []Cylinder{
			{
				Description:    "source",
				CylinderVolume: CylinderVolume(cylinderConfiguration.SourceCylinderVolume),
				Pressure:       cylinderConfiguration.SourceCylinderPressure,
			},
		}
	}
	if cylinderConfiguration.DestinationCylinderIsTwinset {
		*destinationCylinders = []Cylinder{
			{
				Description:    "left",
				CylinderVolume: CylinderVolume(cylinderConfiguration.DestinationCylinderVolume / 2),
				Pressure:       cylinderConfiguration.DestinationCylinderPressure,
			},
			{
				Description:    "right",
				CylinderVolume: CylinderVolume(cylinderConfiguration.DestinationCylinderVolume / 2),
				Pressure:       cylinderConfiguration.DestinationCylinderPressure,
			},
		}
	} else {
		*destinationCylinders = []Cylinder{
			{
				Description:    "destination",
				CylinderVolume: CylinderVolume(cylinderConfiguration.DestinationCylinderVolume),
				Pressure:       cylinderConfiguration.DestinationCylinderPressure,
			},
		}
	}
}

// CylinderSummary has information about the end result of gas transfers
type CylinderSummary struct {
	Description                  string
	DestinationCylinderGasVolume GasVolume
	DestinationCylinderGasWeight GasWeight
	DestinationCylinderPressure  PressureBar
	SourceCylinderGasVolume      GasVolume
	SourceCylinderPressure       PressureBar
	SourceCylinderGasWeight      GasWeight
}

func equalizeAndReport(cylinderConfiguration CylinderConfiguration, gasSystem GasSystem, gasComposition GasComposition, temperature Temperature, verbose bool, debug bool, printSourceSummary bool) CylinderSummary {
	var sourceCylinders CylinderList
	var destinationCylinders CylinderList
	initializeCylinders(cylinderConfiguration, &sourceCylinders, &destinationCylinders)
	if printSourceSummary {
		sourceCylinderGasVolume := sourceCylinders.TotalGasVolume(gasSystem, gasComposition, temperature)
		destinationCylinderGasVolume := destinationCylinders.TotalGasVolume(gasSystem, gasComposition, temperature)
		if verbose {
			fmt.Println("Before any transfers:")
			fmt.Println("Source cylinders:", sourceCylinderGasVolume, "l of gas, pressure", cylinderConfiguration.SourceCylinderPressure, "bar")
			fmt.Println("Destination cylinders:", destinationCylinderGasVolume, "l of gas, pressure", cylinderConfiguration.DestinationCylinderPressure, "bar")
			fmt.Println()
		}
	}

	var description string
	if cylinderConfiguration.DestinationCylinderIsTwinset && cylinderConfiguration.SourceCylinderIsTwinset {
		description = "both manifolds closed"
	} else if cylinderConfiguration.DestinationCylinderIsTwinset {
		description = "destination manifold closed"
	} else if cylinderConfiguration.SourceCylinderIsTwinset {
		description = "source manifold closed"
	} else {
		description = "all manifolds open"
	}
	fmt.Println("Equalizing with", description)
	stepI := 0
	for sourceI := range sourceCylinders {
		for destinationI := range destinationCylinders {
			stepI++
			destinationCylinderGasVolumeBefore := destinationCylinders[destinationI].GasVolume(gasSystem, gasComposition, temperature)
			destinationCylinders[destinationI].Equalize(&sourceCylinders[sourceI], gasSystem, gasComposition, temperature, verbose, debug)
			if verbose {
				fmt.Printf("Step %d: from %s to %s; transferred %.0fl of gas\n", stepI, sourceCylinders[sourceI].Description, destinationCylinders[destinationI].Description, destinationCylinders[destinationI].GasVolume(gasSystem, gasComposition, temperature)-destinationCylinderGasVolumeBefore)
			}
		}
	}
	destinationCylinderPointers := make([]*Cylinder, len(destinationCylinders))
	for destinationI := range destinationCylinders {
		destinationCylinderPointers[destinationI] = &destinationCylinders[destinationI]
	}
	Equalize(destinationCylinderPointers, gasSystem, gasComposition, temperature, verbose, debug)
	if debug {
		fmt.Println("Source cylinders gas volume:", sourceCylinders.TotalGasVolume(gasSystem, gasComposition, temperature))
		fmt.Println("Destination cylinders gas volume:", destinationCylinders.TotalGasVolume(gasSystem, gasComposition, temperature))
	}
	sourceCylinderGasVolume := sourceCylinders.TotalGasVolume(gasSystem, gasComposition, temperature)
	sourceCylinderPressure := PressureFromVolumes(sourceCylinderGasVolume, sourceCylinders.TotalVolume())
	destinationCylinderGasVolume := destinationCylinders.TotalGasVolume(gasSystem, gasComposition, temperature)
	destinationCylinderPressure := PressureFromVolumes(destinationCylinderGasVolume, destinationCylinders.TotalVolume())
	fmt.Printf("Source cylinders: %.0fl, %.0fbar\n", sourceCylinderGasVolume, sourceCylinderPressure)
	fmt.Printf("Destination cylinders: %.0fl, %.0fbar\n", destinationCylinderGasVolume, destinationCylinderPressure)
	fmt.Println()
	return CylinderSummary{
		Description:                  description,
		DestinationCylinderGasVolume: destinationCylinderGasVolume,
		DestinationCylinderGasWeight: destinationCylinders.TotalGasWeight(gasComposition, temperature),
		DestinationCylinderPressure:  destinationCylinderPressure,
		SourceCylinderGasVolume:      sourceCylinderGasVolume,
		SourceCylinderGasWeight:      sourceCylinders.TotalGasWeight(gasComposition, temperature),
		SourceCylinderPressure:       sourceCylinderPressure,
	}
}
func printSummaries(cylinderSummaries []CylinderSummary, verbose bool) {
	var worstDestinationPressure PressureBar
	for _, cylinderSummary := range cylinderSummaries {
		if cylinderSummary.DestinationCylinderPressure < worstDestinationPressure || worstDestinationPressure == 0 {
			worstDestinationPressure = cylinderSummary.DestinationCylinderPressure
		}
	}

	fmt.Printf("%30s src bar  src l  dst bar  dst l improvement\n", "")
	for _, cylinderSummary := range cylinderSummaries {
		if cylinderSummary.Description == "" {
			continue
		}
		fmt.Printf("%30s %7.0f %6.0f %8.0f %6.0f %10.2f%%\n", cylinderSummary.Description, cylinderSummary.SourceCylinderPressure, cylinderSummary.SourceCylinderGasVolume, cylinderSummary.DestinationCylinderPressure, cylinderSummary.DestinationCylinderGasVolume, 100*(cylinderSummary.DestinationCylinderPressure-worstDestinationPressure)/worstDestinationPressure)
		if verbose {
			fmt.Printf("                            Gas weight %6.0fg         %6.0fg\n", cylinderSummary.SourceCylinderGasWeight, cylinderSummary.DestinationCylinderGasWeight)
		}
	}
}

func main() {
	var verboseFlag = flag.Bool("verbose", false, "Print detailed information")
	var debugFlag = flag.Bool("debug", false, "Print debug information")
	var sourceCylinderVolumeFlag = flag.Float64("source-cylinder-volume", 24, "Source cylinder volume in liters")
	var useIdealGasFlag = flag.Bool("use-ideal-gas", false, "Use ideal gas equations instead of Van der Waals")
	var destinationCylinderVolumeFlag = flag.Float64("destination-cylinder-volume", 24, "Destination cylinder volume in liters")
	var sourceCylinderPressureFlag = flag.Float64("source-cylinder-pressure", 232, "Source cylinder pressure in bar")
	var destinationCylinderPressureFlag = flag.Float64("destination-cylinder-pressure", 100, "Destination cylinder pressure")
	var sourceCylinderIsTwinsetFlag = flag.Bool("source-cylinder-twinset", false, "Source cylinder is a twinset with a closeable manifold")
	var destinationCylinderIsTwinsetFlag = flag.Bool("destination-cylinder-twinset", false, "Destination cylinder is a twinset with a closeable manifold")
	var temperatureFlag = flag.Float64("temperature", 20.0, "Gas temperature for Van der Waals equation (celsius)")
	var heliumPercentFlag = flag.Float64("helium", 0.0, "Percentage of helium")
	var oxygenPercentFlag = flag.Float64("oxygen", 0.21, "Percentage of oxygen")
	var neonPercentFlag = flag.Float64("neon", 0, "Percentage of neon")
	var argonPercentFlag = flag.Float64("argon", 0, "Percentage of argon")
	var hydrogenPercentFlag = flag.Float64("hydrogen", 0, "Percentage of hydrogen")
	flag.Parse()

	if *temperatureFlag < -30 || *temperatureFlag > 80 {
		println("Invalid temperature. Must be >-30 and <80")
		os.Exit(1)
	}

	gasSum := *heliumPercentFlag + *oxygenPercentFlag + *neonPercentFlag + *argonPercentFlag + *hydrogenPercentFlag
	if gasSum > 1.0 {
		println("Defined gases must not exceed 100% (1.0)")
		os.Exit(11)
	}
	nitrogenPercent := 1.0 - gasSum
	gasComposition := GasComposition{
		Argon:    *argonPercentFlag,
		Helium:   *heliumPercentFlag,
		Hydrogen: *hydrogenPercentFlag,
		Neon:     *neonPercentFlag,
		Nitrogen: nitrogenPercent,
		Oxygen:   *oxygenPercentFlag,
	}

	if *destinationCylinderPressureFlag > 350 || *destinationCylinderPressureFlag < 0 {
		println("Invalid destination cylinder pressure; must be >= 0 and <=350")
		os.Exit(1)
	}
	if *sourceCylinderPressureFlag > 350 || *sourceCylinderPressureFlag <= 0 {
		println("Invalid source cylinder pressure; must be > 0 and <=350")
		os.Exit(1)
	}
	if *sourceCylinderPressureFlag < *destinationCylinderPressureFlag {
		println("Source pressure must be higher than destination pressure")
		os.Exit(1)
	}
	if *destinationCylinderVolumeFlag <= 0 || *destinationCylinderVolumeFlag > 1000 {
		println("Destination cylinder volume size must be greater than 0 and less than 1000")
		os.Exit(1)
	}
	if *sourceCylinderVolumeFlag <= 0 || *sourceCylinderVolumeFlag > 1000 {
		println("Source cylinder volume size must be greater than 0 and less than 1000")
		os.Exit(1)
	}
	temperature := Temperature(*temperatureFlag + 273.15)
	var gasSystem GasSystem
	if *useIdealGasFlag {
		gasSystem = IdealGas
	} else {
		gasSystem = VanDerWaals
	}

	cylinderConfiguration := CylinderConfiguration{
		DestinationCylinderIsTwinset: *destinationCylinderIsTwinsetFlag,
		DestinationCylinderPressure:  PressureBar(*destinationCylinderPressureFlag),
		DestinationCylinderVolume:    CylinderVolume(*destinationCylinderVolumeFlag),
		SourceCylinderIsTwinset:      *sourceCylinderIsTwinsetFlag,
		SourceCylinderPressure:       PressureBar(*sourceCylinderPressureFlag),
		SourceCylinderVolume:         CylinderVolume(*sourceCylinderVolumeFlag),
	}
	cylinderSummaries := make([]CylinderSummary, 4)
	a := 0

	cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, gasSystem, gasComposition, temperature, *verboseFlag, *debugFlag, true)
	a++
	if *sourceCylinderIsTwinsetFlag {
		cylinderConfiguration.SourceCylinderIsTwinset = false
		cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, gasSystem, gasComposition, temperature, *verboseFlag, *debugFlag, true)
		a++
		cylinderConfiguration.SourceCylinderIsTwinset = true
	}
	if *destinationCylinderIsTwinsetFlag {
		cylinderConfiguration.DestinationCylinderIsTwinset = false
		cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, gasSystem, gasComposition, temperature, *verboseFlag, *debugFlag, true)
		a++
		cylinderConfiguration.DestinationCylinderIsTwinset = true
	}
	if *destinationCylinderIsTwinsetFlag || *sourceCylinderIsTwinsetFlag {
		cylinderConfiguration.DestinationCylinderIsTwinset = false
		cylinderConfiguration.SourceCylinderIsTwinset = false
		cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, gasSystem, gasComposition, temperature, *verboseFlag, *debugFlag, true)
		a++
	}
	printSummaries(cylinderSummaries, *verboseFlag)

}
