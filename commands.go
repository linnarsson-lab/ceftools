package ceftools

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func CmdAggregate(mean bool, cv bool, stdev bool, maxValue bool, minValue bool, noise string, bycol bool) error {
	// Read the input
	cef, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}

	if mean {
		meanAttr := Attribute{"Mean", make([]string, cef.Rows)}
		cef.RowAttributes = append(cef.RowAttributes, meanAttr)
		for i := 0; i < cef.Rows; i++ {
			row := cef.GetRow(i)
			sum := 0.0
			for j := 0; j < len(row); j++ {
				sum = sum + float64(row[j])
			}
			sum = sum / float64(len(row))
			meanAttr.Values[i] = strconv.FormatFloat(sum, 'f', -1, 64)
		}
	}

	if cv {
		cvAttr := Attribute{"CV", make([]string, cef.Rows)}
		cef.RowAttributes = append(cef.RowAttributes, cvAttr)
		for i := 0; i < cef.Rows; i++ {
			row := cef.GetRow(i)
			mean := 0.0
			for j := 0; j < len(row); j++ {
				mean = mean + float64(row[j])
			}
			mean = mean / float64(len(row))

			stdev := 0.0
			for j := 0; j < len(row); j++ {
				stdev += (float64(row[j]) - mean) * (float64(row[j]) - mean)
			}
			if mean == 0 {
				mean = 1
			}
			cvAttr.Values[i] = strconv.FormatFloat(math.Sqrt(stdev/float64(len(row)))/mean, 'f', -1, 64)
		}
	}

	if noise != "" {
		// Calculate means and CVs
		log2_m := make([]float64, cef.Rows)
		log2_cv := make([]float64, cef.Rows)
		for i := 0; i < cef.Rows; i++ {
			row := cef.GetRow(i)
			mu := 0.0
			for j := 0; j < len(row); j++ {
				mu = mu + float64(row[j])
			}
			mu = mu / float64(cef.Columns)
			log2_m[i] = math.Log2(mu)

			sd := 0.0
			for j := 0; j < len(row); j++ {
				sd += (float64(row[j]) - mu) * (float64(row[j]) - mu)
			}
			if mu > 0.0 {
				log2_cv[i] = math.Log2(math.Sqrt(sd/float64(cef.Columns)) / mu)
			} else {
				log2_cv[i] = 0.0
			}
		}

		// Least squares fit to the CV-vs-mean curve; below is the Python code from Gioele

		// fun = lambda x, log2_m, log2_cv: sum(abs( log2( (2.**log2_m)**(-x[0])+x[1]) - log2_cv ))
		// #Fit using Nelder-Mead algorythm
		// optimization =  minimize(fun, x0, args=(log2_m,log2_cv), method='Nelder-Mead')
		// params = optimization.x
		// #The fitted function
		// fitted_fun = lambda log_mu: log2( (2.**log_mu)**(-params[0]) + params[1])
		// # Score is the relative position with respect of the fitted curve
		// score = log2(cv) - fitted_fun(log2(mu))

		// Two-parameter model, least absolute error
		if noise == "std" {
			f := func(x []float64) float64 {
				sum := 0.0
				for ix := 0; ix < len(log2_m); ix++ {
					if log2_cv[ix] == 0.0 {
						continue
					}
					sum += math.Abs(math.Log2(math.Pow(math.Pow(2, log2_m[ix]), -x[0])+x[1]) - log2_cv[ix])
				}
				return sum
			}

			// Minimize the error
			r, _, _ := Minimize(f, [][]float64{[]float64{0.5, 0.2}, []float64{0.6, 0.2}, []float64{0.55, 0.75}}, nil)

			// Calculate offset from the fit
			f2 := func(log_mu float64) float64 {
				return math.Log2(math.Pow(math.Pow(2, log_mu), -r[0]) + r[1])
			}
			fmt.Fprintf(os.Stderr, "log2(CV) = log2(mean^-%.2f) + %.2f\n", r[0], r[1])

			noiseAttr := Attribute{"Noise", make([]string, cef.Rows)}
			for i := 0; i < cef.Rows; i++ {
				noiseAttr.Values[i] = strconv.FormatFloat(log2_cv[i]-f2(log2_m[i]), 'f', -1, 64)
			}
			cef.RowAttributes = append(cef.RowAttributes, noiseAttr)
		} else if noise == "bands" {
			// One-parameter model, minimum band ratio
			const SLOPE = 0.54
			f := func(x []float64) float64 {
				sum := 0.0
				count := 0.0
				for ix := 0; ix < len(log2_m); ix++ {
					if log2_m[ix] == 0.0 || log2_cv[ix] == 0.0 {
						continue
					}
					dist := math.Log2(math.Pow(math.Pow(2, log2_m[ix]), -SLOPE)+x[0]) - log2_cv[ix]
					if dist > 0 {
						sum++
					}
					count++
				}
				//				fmt.Fprintf(os.Stderr, "f(%v) = %v\n", x[0], innersum/outersum)
				return math.Abs(sum - count/2)
			}

			// Minimize the error
			r, _, _ := Minimize(f, [][]float64{[]float64{0.4}, []float64{0.6}}, nil)

			// Calculate offset from the fit
			f2 := func(log_mu float64) float64 {
				return math.Log2(math.Pow(math.Pow(2, log_mu), -SLOPE) + r[0])
			}
			fmt.Fprintf(os.Stderr, "log2(CV) = log2(mean^-%.2f) + %.2f\n", SLOPE, r[0])

			noiseAttr := Attribute{"Noise", make([]string, cef.Rows)}
			for i := 0; i < cef.Rows; i++ {
				//			noiseAttr.Values[i] = strconv.FormatFloat(log2_cv[i]-f2(log2_m[i]), 'f', -1, 64)
				noiseAttr.Values[i] = strconv.FormatFloat(log2_cv[i]-f2(log2_m[i]), 'f', -1, 64)
			}
			cef.RowAttributes = append(cef.RowAttributes, noiseAttr)
		}
	}

	if stdev {
		stdevAttr := Attribute{"Stdev", make([]string, cef.Rows)}
		cef.RowAttributes = append(cef.RowAttributes, stdevAttr)
		for i := 0; i < cef.Rows; i++ {
			row := cef.GetRow(i)
			mean := 0.0
			for j := 0; j < len(row); j++ {
				mean = mean + float64(row[j])
			}
			mean = mean / float64(len(row))

			stdev := 0.0
			for j := 0; j < len(row); j++ {
				stdev += (float64(row[j]) - mean) * (float64(row[j]) - mean)
			}
			stdevAttr.Values[i] = strconv.FormatFloat(math.Sqrt(stdev/float64(len(row))), 'f', -1, 64)
		}
	}

	if maxValue {
		maxAttr := Attribute{"Max", make([]string, cef.Rows)}
		cef.RowAttributes = append(cef.RowAttributes, maxAttr)
		for i := 0; i < cef.Rows; i++ {
			row := cef.GetRow(i)
			max := float64(row[0])
			for j := 0; j < len(row); j++ {
				max = math.Max(max, float64(row[j]))
			}
			maxAttr.Values[i] = strconv.FormatFloat(max, 'f', -1, 64)
		}
	}
	if minValue {
		minAttr := Attribute{"Min", make([]string, cef.Rows)}
		cef.RowAttributes = append(cef.RowAttributes, minAttr)
		for i := 0; i < cef.Rows; i++ {
			row := cef.GetRow(i)
			min := float64(row[0])
			for j := 0; j < len(row); j++ {
				min = math.Min(min, float64(row[j]))
			}
			minAttr.Values[i] = strconv.FormatFloat(min, 'f', -1, 64)
		}
	}
	// Write the CEB file
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdRename(attr string, bycol bool) error {
	temp := strings.Split(attr, "=")
	if len(temp) != 2 {
		return errors.New("Invalid rename (should be --attr old=new)")
	}

	// Read the input
	cef, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}

	// Rename
	for i := 0; i < len(cef.RowAttributes); i++ {
		if cef.RowAttributes[i].Name == temp[0] {
			cef.RowAttributes[i].Name = temp[1]
			break // Rename only the first instance if there are multiple with same name
		}
	}

	// Write the CEB file
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdSort(sort_by string, sort_numerical bool, reverse bool, bycol bool) error {
	// Read the input
	cef, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}
	var result *Cef
	if strings.Contains(sort_by, "=") || sort_numerical { // If there is a '=' we're sorting on column values, so numerical by definition
		result, err = cef.SortNumerical(sort_by, reverse)
	} else {
		result, err = cef.SortByRowAttribute(sort_by, reverse)
	}
	if err != nil {
		return err
	}
	// Write the CEB file
	if err := Write(result, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdSPIN(corrfile string, bycol bool) error {
	// Read the input
	cef, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}

	cef.SPIN(10, int(math.Log2(float64(cef.Rows))+1), 2*float64(cef.Rows), 0.5, corrfile)

	// Write the CEF file
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdSelect(selector string, bycol bool, except bool) error {
	// Read the input
	cef, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}

	// Parse the selector
	av := strings.Split(selector, "=")
	if len(av) != 2 {
		return errors.New("Invalid --where clause (should be 'attr=value')")
	}
	attr := av[0]
	value := av[1]

	// Make empty slices reusing the existing storage
	tempMatrix := cef.Matrix[:0]
	tempRowAttrValues := make([]Attribute, len(cef.RowAttributes))
	attrIndex := -1
	for i := 0; i < len(tempRowAttrValues); i++ {
		tempRowAttrValues[i] = Attribute{cef.RowAttributes[i].Name, cef.RowAttributes[i].Values[:0]}
		if cef.RowAttributes[i].Name == attr {
			attrIndex = i
		}
	}
	if attrIndex == -1 {
		return errors.New("Attribute not found when attempting to select")
	}

	// Scan all rows for matches and append them
	nSelected := 0
	for i := 0; i < cef.Rows; i++ {
		if cef.RowAttributes[attrIndex].Values[i] == value {
			nSelected++
			tempMatrix = append(tempMatrix, cef.GetRow(i)...)
			for j := 0; j < len(cef.RowAttributes); j++ {
				tempRowAttrValues[j].Values = append(tempRowAttrValues[j].Values, cef.RowAttributes[j].Values[i])
			}
		}
	}

	// Replace the orginals with the filtered copies
	cef.Matrix = tempMatrix
	for i := 0; i < len(cef.RowAttributes); i++ {
		cef.RowAttributes[i] = tempRowAttrValues[i]
	}
	cef.Rows = nSelected

	// Write the CEB file
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdSelectRange(from int, to int, bycol bool, except bool) error {
	// Read the input
	cef, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}
	if to == -1 {
		to = cef.Rows
	}
	if to < from {
		temp := to
		from = to
		to = temp
	}
	if from < 1 {
		from = 1
	}
	if from > cef.Rows {
		from = cef.Rows
	}
	if to < 1 {
		to = 1
	}
	if to > cef.Rows {
		to = cef.Rows
	}

	// Truncate the matrix
	if except {
		// Set the matrix to the first segment
		temp := cef.Matrix[:(from-1)*cef.Columns]
		// Add the remaining segment
		for i := to; i < cef.Rows; i++ {
			temp = append(temp, cef.GetRow(i)...)
		}
		cef.Matrix = temp

		// And same for row attributes
		for i := 0; i < len(cef.RowAttributes); i++ {
			temp2 := cef.RowAttributes[i].Values[:(from - 1)]
			temp2 = append(temp2, cef.RowAttributes[i].Values[to:]...)
			cef.RowAttributes[i].Values = temp2
		}
		cef.Rows = cef.Rows - (to - from + 1)
	} else {
		cef.Rows = to - from + 1
		cef.Matrix = cef.Matrix[(from-1)*cef.Columns : to*cef.Columns]
		for i := 0; i < len(cef.RowAttributes); i++ {
			cef.RowAttributes[i].Values = cef.RowAttributes[i].Values[from-1 : to]
		}
	}

	// Write the CEB file
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdJoin(other string, on string, bycol bool) error {
	// Read the input
	left, err := Read(os.Stdin, bycol)
	if err != nil {
		return err
	}
	// Read the right (to be joined)
	f, err := os.Open(other)
	if err != nil {
		return err
	}
	defer f.Close()
	right, err := Read(f, bycol)
	if err != nil {
		return err
	}
	// Perform the join
	attrs := strings.Split(on, "=")
	if len(attrs) != 2 {
		return errors.New("--on 'attr1=attr2' was incorrectly specified")
	}
	cef, err := left.Join(right, attrs[0], attrs[1])
	if err != nil {
		return err
	}
	// Write the CEB file
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdAdd(attr string, header string, bycol bool) error {
	// Read the input
	var cef, err = Read(os.Stdin, bycol)
	if err != nil {
		return err
	}

	if attr != "" {
		nv := strings.Split(attr, "=")
		if len(nv) != 2 {
			return errors.New("Invalid 'Name=value' string when attempting to add attribute")
		}
		newAttr := Attribute{nv[0], make([]string, cef.Rows)}
		for i := 0; i < cef.Rows; i++ {
			if nv[1] == "(row)" {
				newAttr.Values[i] = strconv.Itoa(int(i + 1))
			} else {
				newAttr.Values[i] = nv[1]
			}
		}
		cef.RowAttributes = append(cef.RowAttributes, newAttr)
	}
	if header != "" {
		nv := strings.Split(header, "=")
		if len(nv) != 2 {
			return errors.New("Invalid 'Name=value' string when attempting to add header")
		}
		newHdr := Header{nv[0], nv[1]}
		cef.Headers = append(cef.Headers, newHdr)
	}

	// Write the result
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}
func CmdDrop(attrs string, headers string, except bool, bycol bool) error {
	// Read the input
	var cef, err = Read(os.Stdin, bycol)
	if err != nil {
		return err
	}
	if attrs != "" {
		dropAttrs(cef, attrs, except)
	}
	if headers != "" {
		dropHeaders(cef, headers, except)
	}

	// Write the result
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func dropAttrs(cef *Cef, attrs string, except bool) {
	// Drop the attributes
	todrop := strings.Split(attrs, ",")
	temp := cef.RowAttributes[:0]
	for _, att := range cef.RowAttributes {
		if contains(todrop, att.Name) == except {
			temp = append(temp, att)
		}
	}
	cef.RowAttributes = temp
}

func dropHeaders(cef *Cef, headers string, except bool) {

	// Drop the attributes
	todrop := strings.Split(headers, ",")
	temp := cef.Headers[:0]
	for _, hdr := range cef.Headers {
		if contains(todrop, hdr.Name) == except {
			temp = append(temp, hdr)
		}
	}
	cef.Headers = temp
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CmdRescale(method string, length_attr string, bycol bool) error {
	// Read the input
	var cef, err = Read(os.Stdin, bycol)
	if err != nil {
		return err
	}

	log_rescale := func(vals []float32) {
		for i := 0; i < len(vals); i++ {
			vals[i] = float32(math.Log10(float64(vals[i] + 1)))
		}
	}
	tpm_rescale := func(vals []float32) {
		sum := float32(0)
		for i := 0; i < len(vals); i++ {
			sum += vals[i]
		}
		if sum != 0 {
			for i := 0; i < len(vals); i++ {
				vals[i] = vals[i] * 1000000 / sum
			}
		}
	}
	rpkm_rescale := func(vals []float32, length float32) {
		sum := float32(0)
		for i := 0; i < len(vals); i++ {
			sum += vals[i]
		}
		if length == 0 {
			length = 1
		}
		if sum != 0 {
			for i := 0; i < len(vals); i++ {
				vals[i] = vals[i] * 1000 / sum / length
			}
		}
	}
	var length []string
	if length_attr != "" {
		for i := 0; i < len(cef.RowAttributes); i++ {
			if cef.RowAttributes[i].Name == length_attr {
				length = cef.RowAttributes[i].Values
			}
		}
		if length == nil {
			return errors.New("Length attribute not found when attempting to rescale by rpkm")
		}
	}
	for i := 0; i < cef.Rows; i++ {
		switch method {
		case "log":
			log_rescale(cef.GetRow(i))
			break
		case "tpm":
			tpm_rescale(cef.GetRow(i))
		case "rpkm":
			bp, err := strconv.Atoi(length[i])
			if err != nil {
				return errors.New("Length attribute was not a valid integer (when attempting to rescale by rpkm)")
			}
			rpkm_rescale(cef.GetRow(i), float32(bp)/1000)
		}
	}

	// Write the result
	if err := Write(cef, os.Stdout, bycol); err != nil {
		return err
	}
	return nil
}

func CmdImportStrt() error {
	cef, err := ReadStrt(os.Stdin, false)
	if err != nil {
		return err
	}

	// Write the CEB file
	if err := Write(cef, os.Stdout, false); err != nil {
		return err
	}
	return nil
}
