package excel

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelManager handles Excel import/export operations
type ExcelManager struct {
	file *excelize.File
}

// NewExcelManager creates a new Excel manager
func NewExcelManager() *ExcelManager {
	return &ExcelManager{
		file: excelize.NewFile(),
	}
}

// NewExcelManagerFromFile creates Excel manager from existing file
func NewExcelManagerFromFile(file *multipart.FileHeader) (*ExcelManager, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}

	return &ExcelManager{file: f}, nil
}

// ExportToExcel exports data to Excel file
func (em *ExcelManager) ExportToExcel(data interface{}, sheetName string, headers []string) error {
	// Create new sheet
	index, err := em.file.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	// Set headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		em.file.SetCellValue(sheetName, cell, header)
	}

	// Convert data to slice
	slice := reflect.ValueOf(data)
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}

	// Write data rows
	for i := 0; i < slice.Len(); i++ {
		row := slice.Index(i)
		if row.Kind() == reflect.Ptr {
			row = row.Elem()
		}

		if row.Kind() != reflect.Struct {
			return fmt.Errorf("slice elements must be structs")
		}

		// Write each field
		for j := 0; j < row.NumField(); j++ {
			field := row.Field(j)
			cell := fmt.Sprintf("%c%d", 'A'+j, i+2) // +2 because row 1 is headers

			// Convert field value to string
			value := em.convertFieldToString(field)
			em.file.SetCellValue(sheetName, cell, value)
		}
	}

	// Set active sheet
	em.file.SetActiveSheet(index)

	return nil
}

// ImportFromExcel imports data from Excel file
func (em *ExcelManager) ImportFromExcel(sheetName string, targetType reflect.Type) (interface{}, error) {
	rows, err := em.file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file must have at least 2 rows (header + data)")
	}

	// Get headers
	headers := rows[0]

	// Create slice of target type
	sliceType := reflect.SliceOf(targetType)
	result := reflect.New(sliceType).Elem()

	// Process data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}

		// Create new instance of target type
		item := reflect.New(targetType).Elem()

		// Map row data to struct fields
		for j, cellValue := range row {
			if j >= len(headers) {
				break
			}

			header := headers[j]
			field := em.findFieldByName(item, header)
			if field.IsValid() && field.CanSet() {
				em.setFieldFromString(field, cellValue)
			}
		}

		result = reflect.Append(result, item)
	}

	return result.Interface(), nil
}

// ExportToCSV exports data to CSV format
func (em *ExcelManager) ExportToCSV(data interface{}, headers []string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write headers
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Convert data to slice
	slice := reflect.ValueOf(data)
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}

	// Write data rows
	for i := 0; i < slice.Len(); i++ {
		row := slice.Index(i)
		if row.Kind() == reflect.Ptr {
			row = row.Elem()
		}

		if row.Kind() != reflect.Struct {
			return fmt.Errorf("slice elements must be structs")
		}

		// Convert struct to string slice
		var record []string
		for j := 0; j < row.NumField(); j++ {
			field := row.Field(j)
			value := em.convertFieldToString(field)
			record = append(record, value)
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// ImportFromCSV imports data from CSV format
func (em *ExcelManager) ImportFromCSV(reader io.Reader, targetType reflect.Type) (interface{}, error) {
	csvReader := csv.NewReader(reader)

	// Read headers
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Create slice of target type
	sliceType := reflect.SliceOf(targetType)
	result := reflect.New(sliceType).Elem()

	// Read data rows
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		if len(record) == 0 {
			continue // Skip empty rows
		}

		// Create new instance of target type
		item := reflect.New(targetType).Elem()

		// Map record data to struct fields
		for j, cellValue := range record {
			if j >= len(headers) {
				break
			}

			header := headers[j]
			field := em.findFieldByName(item, header)
			if field.IsValid() && field.CanSet() {
				em.setFieldFromString(field, cellValue)
			}
		}

		result = reflect.Append(result, item)
	}

	return result.Interface(), nil
}

// Save saves the Excel file
func (em *ExcelManager) Save(filename string) error {
	return em.file.SaveAs(filename)
}

// WriteToWriter writes Excel file to writer
func (em *ExcelManager) WriteToWriter(writer io.Writer) error {
	return em.file.Write(writer)
}

// GetSheetNames returns all sheet names
func (em *ExcelManager) GetSheetNames() []string {
	return em.file.GetSheetList()
}

// GetSheetData returns data from a specific sheet
func (em *ExcelManager) GetSheetData(sheetName string) ([][]string, error) {
	return em.file.GetRows(sheetName)
}

// convertFieldToString converts a field value to string
func (em *ExcelManager) convertFieldToString(field reflect.Value) string {
	if !field.IsValid() {
		return ""
	}

	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(field.Bool())
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			return field.Interface().(time.Time).Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%v", field.Interface())
	default:
		return fmt.Sprintf("%v", field.Interface())
	}
}

// findFieldByName finds a struct field by name (case-insensitive)
func (em *ExcelManager) findFieldByName(item reflect.Value, name string) reflect.Value {
	itemType := item.Type()

	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		fieldName := field.Name

		// Check for json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			jsonTag = strings.Split(jsonTag, ",")[0]
			if jsonTag == name {
				return item.Field(i)
			}
		}

		// Check for excel tag
		if excelTag := field.Tag.Get("excel"); excelTag != "" {
			if excelTag == name {
				return item.Field(i)
			}
		}

		// Check field name (case-insensitive)
		if strings.EqualFold(fieldName, name) {
			return item.Field(i)
		}
	}

	return reflect.Value{}
}

// setFieldFromString sets a field value from string
func (em *ExcelManager) setFieldFromString(field reflect.Value, value string) {
	if !field.IsValid() || !field.CanSet() {
		return
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		}
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			// Try different time formats
			formats := []string{
				"2006-01-02 15:04:05",
				"2006-01-02",
				"01/02/2006",
				"2006-01-02T15:04:05Z",
			}

			for _, format := range formats {
				if timeVal, err := time.Parse(format, value); err == nil {
					field.Set(reflect.ValueOf(timeVal))
					break
				}
			}
		}
	}
}

// ValidationError represents a validation error for Excel import
type ValidationError struct {
	Row    int    `json:"row"`
	Column string `json:"column"`
	Value  string `json:"value"`
	Error  string `json:"error"`
}

// ValidateData validates imported data
func (em *ExcelManager) ValidateData(data interface{}, validators map[string]func(string) error) []ValidationError {
	var errors []ValidationError

	slice := reflect.ValueOf(data)
	if slice.Kind() != reflect.Slice {
		return errors
	}

	for i := 0; i < slice.Len(); i++ {
		row := slice.Index(i)
		if row.Kind() == reflect.Ptr {
			row = row.Elem()
		}

		if row.Kind() != reflect.Struct {
			continue
		}

		rowType := row.Type()
		for j := 0; j < row.NumField(); j++ {
			field := row.Field(j)
			fieldType := rowType.Field(j)

			// Get field name for validation
			fieldName := fieldType.Name
			if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
				fieldName = strings.Split(jsonTag, ",")[0]
			}

			// Validate field if validator exists
			if validator, exists := validators[fieldName]; exists {
				value := em.convertFieldToString(field)
				if err := validator(value); err != nil {
					errors = append(errors, ValidationError{
						Row:    i + 2, // +2 because Excel rows start from 1 and row 1 is header
						Column: fieldName,
						Value:  value,
						Error:  err.Error(),
					})
				}
			}
		}
	}

	return errors
}
