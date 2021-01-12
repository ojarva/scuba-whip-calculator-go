package main

import (
	"flag"
	"fmt"
	"os"
)

// Equalize equalizes all input cylinders
func Equalize(cylinders []*Cylinder, verbose bool, debug bool) {
	var totalGasVolume float64
	var totalVolume float64
	for i := range cylinders {
		totalGasVolume += cylinders[i].GasVolume()
		totalVolume += cylinders[i].CylinderVolume
	}
	pressureAfterEqualize := totalGasVolume / totalVolume
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
func (c1 Cylinder) GasVolume() float64 {
	return c1.CylinderVolume * c1.Pressure
}

// Equalize equalizes two cylinders
func (c1 *Cylinder) Equalize(c2 *Cylinder, verbose bool, debug bool) {
	listOfCylinders := []*Cylinder{c1, c2}
	Equalize(listOfCylinders, verbose, debug)
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
func (cl CylinderList) TotalGasVolume() float64 {
	var totalGasVolume float64
	for _, cylinder := range cl {
		totalGasVolume += cylinder.GasVolume()
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

func equalizeAndReport(cylinderConfiguration CylinderConfiguration, verbose bool, debug bool, printSourceSummary bool) CylinderSummary {
	var sourceCylinders CylinderList
	var destinationCylinders CylinderList
	initializeCylinders(cylinderConfiguration, &sourceCylinders, &destinationCylinders)
	if printSourceSummary {
		sourceCylinderGasVolume := sourceCylinders.TotalGasVolume()
		destinationCylinderGasVolume := destinationCylinders.TotalGasVolume()
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
			destinationCylinderGasVolumeBefore := destinationCylinders[destinationI].GasVolume()
			destinationCylinders[destinationI].Equalize(&sourceCylinders[sourceI], verbose, debug)
			if verbose {
				fmt.Printf("Step %d: from %s to %s; transferred %.0fl of gas\n", stepI, sourceCylinders[sourceI].Description, destinationCylinders[destinationI].Description, destinationCylinders[destinationI].GasVolume()-destinationCylinderGasVolumeBefore)
			}
		}
	}
	destinationCylinderPointers := make([]*Cylinder, len(destinationCylinders))
	for destinationI := range destinationCylinders {
		destinationCylinderPointers[destinationI] = &destinationCylinders[destinationI]
	}
	Equalize(destinationCylinderPointers, verbose, debug)
	if debug {
		fmt.Println("Source cylinders gas volume:", sourceCylinders.TotalGasVolume())
		fmt.Println("Destination cylinders gas volume:", destinationCylinders.TotalGasVolume())
	}
	sourceCylinderGasVolume := sourceCylinders.TotalGasVolume()
	sourceCylinderPressure := sourceCylinderGasVolume / sourceCylinders.TotalVolume()
	destinationCylinderGasVolume := destinationCylinders.TotalGasVolume()
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
	var destinationCylinderVolumeFlag = flag.Float64("destination-cylinder-volume", 24, "Destination cylinder volume in liters")
	var sourceCylinderPressureFlag = flag.Float64("source-cylinder-pressure", 232, "Source cylinder pressure in bar")
	var destinationCylinderPressureFlag = flag.Float64("destination-cylinder-pressure", 100, "Destination cylinder pressure")
	var sourceCylinderIsTwinsetFlag = flag.Bool("source-cylinder-twinset", false, "Source cylinder is a twinset with a closeable manifold")
	var destinationCylinderIsTwinsetFlag = flag.Bool("destination-cylinder-twinset", false, "Destination cylinder is a twinset with a closeable manifold")
	flag.Parse()

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

	cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, *verboseFlag, *debugFlag, true)
	a++
	if *sourceCylinderIsTwinsetFlag {
		cylinderConfiguration.SourceCylinderIsTwinset = false
		cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, *verboseFlag, *debugFlag, false)
		a++
		cylinderConfiguration.SourceCylinderIsTwinset = true
	}
	if *destinationCylinderIsTwinsetFlag {
		cylinderConfiguration.DestinationCylinderIsTwinset = false
		cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, *verboseFlag, *debugFlag, false)
		a++
		cylinderConfiguration.DestinationCylinderIsTwinset = true
	}
	if *destinationCylinderIsTwinsetFlag || *sourceCylinderIsTwinsetFlag {
		cylinderConfiguration.DestinationCylinderIsTwinset = false
		cylinderConfiguration.SourceCylinderIsTwinset = false
		cylinderSummaries[a] = equalizeAndReport(cylinderConfiguration, *verboseFlag, *debugFlag, false)
		a++
	}
	printSummaries(cylinderSummaries)

}
