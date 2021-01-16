package main

import (
	"flag"
	"fmt"
	"math"
	"os"
)

type GasSystem int

const (
	IdealGas GasSystem = iota
	VanDerWaals
)

type Gas int

const (
	Helium Gas = iota
	Oxygen
	Nitrogen
	Argon
	Neon
	Hydrogen
)

type VanDerWaalsConstant struct {
	A float64
	B float64
}

var VanDerWaalsConstants = map[Gas]VanDerWaalsConstant{
	Helium:   VanDerWaalsConstant{0.0346, 0.0238},
	Oxygen:   VanDerWaalsConstant{1.382, 0.03186},
	Nitrogen: VanDerWaalsConstant{1.370, 0.0387},
	Argon:    VanDerWaalsConstant{1.355, 0.03201},
	Neon:     VanDerWaalsConstant{0.2135, 0.01709},
	Hydrogen: VanDerWaalsConstant{0.2476, 0.02661},
}

type GasComposition struct {
	Helium   float64
	Oxygen   float64
	Nitrogen float64
	Argon    float64
	Neon     float64
	Hydrogen float64
}

// Equalize equalizes all input cylinders
func Equalize(cylinders []*Cylinder, gasSystem GasSystem, gasComposition GasComposition, temperature float64, verbose bool, debug bool) {
	var totalVolume float64
	var pressureAfterEqualize float64
	if gasSystem == IdealGas {
		var totalGasVolume float64
		for i := range cylinders {
			totalGasVolume += cylinders[i].GasVolume(gasSystem, gasComposition, temperature)
			totalVolume += cylinders[i].CylinderVolume
		}
		pressureAfterEqualize = totalGasVolume / totalVolume
	} else {
		var totalMoles float64
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
	DestinationCylinderPressure  float64
	DestinationCylinderVolume    float64
	SourceCylinderIsTwinset      bool
	SourceCylinderPressure       float64
	SourceCylinderVolume         float64
}

// Cylinder represents a single cylinder and gas it contains
type Cylinder struct {
	Description    string
	CylinderVolume float64
	Pressure       float64
}

// GasVolume returns amount of gas in the cylinder
func (c1 Cylinder) GasVolume(gasSystem GasSystem, gasComposition GasComposition, temperature float64) float64 {
	if gasSystem == IdealGas {
		return c1.CylinderVolume * c1.Pressure
	} else {
		return gasCompositionToMoles(c1.CylinderVolume, c1.Pressure, temperature, gasComposition) * 22.4
	}
}

// Equalize equalizes two cylinders
func (c1 *Cylinder) Equalize(c2 *Cylinder, gasSystem GasSystem, gasComposition GasComposition, temperature float64, verbose bool, debug bool) {
	listOfCylinders := []*Cylinder{c1, c2}
	Equalize(listOfCylinders, gasSystem, gasComposition, temperature, verbose, debug)
}

func (c1 *Cylinder) Moles(temperature float64, gasComposition GasComposition) float64 {
	return gasCompositionToMoles(c1.CylinderVolume, c1.Pressure, temperature, gasComposition)
}

func gasCompositionToMoles(V float64, P float64, temperature float64, gasComposition GasComposition) float64 {
	return GasToMoles(V, gasComposition.Argon*P, VanDerWaalsConstants[Argon], temperature) + GasToMoles(V, gasComposition.Helium*P, VanDerWaalsConstants[Helium], temperature) + GasToMoles(V, gasComposition.Hydrogen*P, VanDerWaalsConstants[Hydrogen], temperature) + GasToMoles(V, gasComposition.Neon*P, VanDerWaalsConstants[Neon], temperature) + GasToMoles(V, gasComposition.Nitrogen*P, VanDerWaalsConstants[Nitrogen], temperature) + GasToMoles(V, gasComposition.Oxygen*P, VanDerWaalsConstants[Oxygen], temperature)
}

func GasToMoles(V float64, P float64, vdwConstants VanDerWaalsConstant, T float64) float64 {
	R := 0.0831
	a := vdwConstants.A
	b := vdwConstants.B

	a2 := math.Pow(a, 2.0)
	a3 := math.Pow(a, 3.0)
	b2 := math.Pow(b, 2.0)
	V2 := math.Pow(V, 2.0)
	V3 := math.Pow(V, 3.0)

	subterm1 := 2*a3*V3 + 18*a2*b2*P*V3 - 9*a2*b*R*T*V3
	subterm2 := 3*a*b*(b*P*V2+R*T*V2) - a2*V2
	subterm3 := math.Pow(
		(subterm1 +
			math.Sqrt(4*math.Pow(subterm2, 3.0)+math.Pow(subterm1, 2.0))),
		(1 / 3.0))
	term1 := 0.26457 * subterm3
	term2 := a * b * subterm3
	term3 := 0.41997 * subterm2
	return term1/(a*b) - term3/term2 + (0.33333*V)/b
}

func cylinderMolesToPressure(V float64, n float64, temperature float64, gasComposition GasComposition) float64 {
	return MolesToPressure(V, n*gasComposition.Argon, temperature, VanDerWaalsConstants[Argon]) + MolesToPressure(V, n*gasComposition.Helium, temperature, VanDerWaalsConstants[Helium]) + MolesToPressure(V, n*gasComposition.Hydrogen, temperature, VanDerWaalsConstants[Hydrogen]) + MolesToPressure(V, n*gasComposition.Neon, temperature, VanDerWaalsConstants[Neon]) + MolesToPressure(V, n*gasComposition.Nitrogen, temperature, VanDerWaalsConstants[Nitrogen]) + MolesToPressure(V, n*gasComposition.Oxygen, temperature, VanDerWaalsConstants[Oxygen])
}

func MolesToPressure(V float64, n float64, T float64, vdwConstants VanDerWaalsConstant) float64 {
	a := vdwConstants.A
	b := vdwConstants.B
	R := 0.0831
	V2 := math.Pow(V, 2.0)
	return n * (-(a*n)/V2 - (R*T)/(b*n-V))
}

// CylinderList is a list of cylinders
type CylinderList []Cylinder

// TotalVolume returns total cylinder volume for all listed cylinders
func (cl CylinderList) TotalVolume() float64 {
	var totalVolume float64
	for _, cylinder := range cl {
		totalVolume += cylinder.CylinderVolume
	}
	return totalVolume
}

// TotalGasVolume returns total gas volume for all listed cylinders
func (cl CylinderList) TotalGasVolume(gasSystem GasSystem, gasComposition GasComposition, temperature float64) float64 {
	var totalGasVolume float64
	for _, cylinder := range cl {
		totalGasVolume += cylinder.GasVolume(gasSystem, gasComposition, temperature)
	}
	return totalGasVolume
}

func initializeCylinders(cylinderConfiguration CylinderConfiguration, sourceCylinders *CylinderList, destinationCylinders *CylinderList) {
	if cylinderConfiguration.SourceCylinderIsTwinset {
		*sourceCylinders = []Cylinder{
			Cylinder{
				Description:    "left",
				CylinderVolume: cylinderConfiguration.SourceCylinderVolume / 2,
				Pressure:       cylinderConfiguration.SourceCylinderPressure,
			},
			Cylinder{
				Description:    "right",
				CylinderVolume: cylinderConfiguration.SourceCylinderVolume / 2,
				Pressure:       cylinderConfiguration.SourceCylinderPressure,
			},
		}
	} else {
		*sourceCylinders = []Cylinder{
			Cylinder{
				Description:    "source",
				CylinderVolume: cylinderConfiguration.SourceCylinderVolume,
				Pressure:       cylinderConfiguration.SourceCylinderPressure,
			},
		}
	}
	if cylinderConfiguration.DestinationCylinderIsTwinset {
		*destinationCylinders = []Cylinder{
			Cylinder{
				Description:    "left",
				CylinderVolume: cylinderConfiguration.DestinationCylinderVolume / 2,
				Pressure:       cylinderConfiguration.DestinationCylinderPressure,
			},
			Cylinder{
				Description:    "right",
				CylinderVolume: cylinderConfiguration.DestinationCylinderVolume / 2,
				Pressure:       cylinderConfiguration.DestinationCylinderPressure,
			},
		}
	} else {
		*destinationCylinders = []Cylinder{
			Cylinder{
				Description:    "destination",
				CylinderVolume: cylinderConfiguration.DestinationCylinderVolume,
				Pressure:       cylinderConfiguration.DestinationCylinderPressure,
			},
		}
	}
}

// CylinderSummary has information about the end result of gas transfers
type CylinderSummary struct {
	Description                  string
	DestinationCylinderGasVolume float64
	DestinationCylinderPressure  float64
	SourceCylinderGasVolume      float64
	SourceCylinderPressure       float64
}

func equalizeAndReport(cylinderConfiguration CylinderConfiguration, gasSystem GasSystem, gasComposition GasComposition, temperature float64, verbose bool, debug bool, printSourceSummary bool) CylinderSummary {
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
	sourceCylinderPressure := sourceCylinderGasVolume / sourceCylinders.TotalVolume()
	destinationCylinderGasVolume := destinationCylinders.TotalGasVolume(gasSystem, gasComposition, temperature)
	destinationCylinderPressure := destinationCylinderGasVolume / destinationCylinders.TotalVolume()
	fmt.Printf("Source cylinders: %.0fl, %.0fbar\n", sourceCylinderGasVolume, sourceCylinderPressure)
	fmt.Printf("Destination cylinders: %.0fl, %.0fbar\n", destinationCylinderGasVolume, destinationCylinderPressure)
	fmt.Println()
	return CylinderSummary{
		Description:                  description,
		DestinationCylinderGasVolume: destinationCylinderGasVolume,
		DestinationCylinderPressure:  destinationCylinderPressure,
		SourceCylinderGasVolume:      sourceCylinderGasVolume,
		SourceCylinderPressure:       sourceCylinderPressure,
	}
}
func printSummaries(cylinderSummaries []CylinderSummary) {
	var worstDestinationPressure float64
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
	temperature := *temperatureFlag + 273.15
	var gasSystem GasSystem
	if *useIdealGasFlag {
		gasSystem = IdealGas
	} else {
		gasSystem = VanDerWaals
	}

	cylinderConfiguration := CylinderConfiguration{
		DestinationCylinderIsTwinset: *destinationCylinderIsTwinsetFlag,
		DestinationCylinderPressure:  *destinationCylinderPressureFlag,
		DestinationCylinderVolume:    *destinationCylinderVolumeFlag,
		SourceCylinderIsTwinset:      *sourceCylinderIsTwinsetFlag,
		SourceCylinderPressure:       *sourceCylinderPressureFlag,
		SourceCylinderVolume:         *sourceCylinderVolumeFlag,
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
	printSummaries(cylinderSummaries)

}
