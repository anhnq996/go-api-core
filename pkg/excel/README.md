# Excel Package

Package excel provides comprehensive Excel and CSV import/export functionality with validation support.

## Features

- **Excel Import/Export**: Full support for .xlsx files
- **CSV Import/Export**: Support for CSV format
- **Data Validation**: Built-in validation with custom validators
- **Struct Mapping**: Automatic mapping between structs and Excel/CSV
- **Flexible Headers**: Support for custom headers and field mapping
- **Error Handling**: Detailed validation errors with row/column information

## Usage

### Basic Excel Export

```go
import "api-core/pkg/excel"

type User struct {
    ID        int       `json:"id" excel:"ID"`
    Name      string    `json:"name" excel:"Name"`
    Email     string    `json:"email" excel:"Email"`
    Age       int       `json:"age" excel:"Age"`
    IsActive  bool      `json:"is_active" excel:"Active"`
    CreatedAt time.Time `json:"created_at" excel:"Created At"`
}

func exportUsers() {
    users := []User{
        {ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30, IsActive: true, CreatedAt: time.Now()},
        {ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25, IsActive: false, CreatedAt: time.Now()},
    }

    // Create Excel manager
    excelManager := excel.NewExcelManager()

    // Define headers
    headers := []string{"ID", "Name", "Email", "Age", "Active", "Created At"}

    // Export to Excel
    err := excelManager.ExportToExcel(users, "Users", headers)
    if err != nil {
        log.Fatal(err)
    }

    // Save file
    err = excelManager.Save("users.xlsx")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Basic Excel Import

```go
func importUsers(file *multipart.FileHeader) ([]User, error) {
    // Create Excel manager from uploaded file
    excelManager, err := excel.NewExcelManagerFromFile(file)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }

    // Import data
    data, err := excelManager.ImportFromExcel("Users", reflect.TypeOf(User{}))
    if err != nil {
        return nil, fmt.Errorf("failed to import data: %w", err)
    }

    return data.([]User), nil
}
```

### CSV Export

```go
func exportUsersToCSV(users []User, writer io.Writer) error {
    excelManager := excel.NewExcelManager()
    headers := []string{"ID", "Name", "Email", "Age", "Active", "Created At"}

    return excelManager.ExportToCSV(users, headers, writer)
}
```

### CSV Import

```go
func importUsersFromCSV(reader io.Reader) ([]User, error) {
    excelManager := excel.NewExcelManager()

    data, err := excelManager.ImportFromCSV(reader, reflect.TypeOf(User{}))
    if err != nil {
        return nil, err
    }

    return data.([]User), nil
}
```

### Data Validation

```go
func validateUsers(users []User) []excel.ValidationError {
    excelManager := excel.NewExcelManager()

    // Define validators
    validators := map[string]func(string) error{
        "Name": func(value string) error {
            if value == "" {
                return fmt.Errorf("name is required")
            }
            return nil
        },
        "Email": func(value string) error {
            if !isValidEmail(value) {
                return fmt.Errorf("invalid email format")
            }
            return nil
        },
        "Age": func(value string) error {
            if value == "" {
                return nil // Optional field
            }
            age, err := strconv.Atoi(value)
            if err != nil {
                return fmt.Errorf("age must be a number")
            }
            if age < 0 || age > 150 {
                return fmt.Errorf("age must be between 0 and 150")
            }
            return nil
        },
    }

    return excelManager.ValidateData(users, validators)
}
```

## Struct Tags

### Excel Tag

Use the `excel` tag to specify the column name in Excel:

```go
type User struct {
    ID   int    `excel:"ID"`
    Name string `excel:"Name"`
    Email string `excel:"Email Address"`
}
```

### JSON Tag

The package also supports `json` tags as fallback:

```go
type User struct {
    ID   int    `json:"id" excel:"ID"`
    Name string `json:"name" excel:"Name"`
}
```

## HTTP Handler Examples

### Export Handler

```go
func ExportUsersHandler(w http.ResponseWriter, r *http.Request) {
    // Get users from database
    users, err := userService.GetAll()
    if err != nil {
        http.Error(w, "Failed to get users", http.StatusInternalServerError)
        return
    }

    // Create Excel manager
    excelManager := excel.NewExcelManager()

    // Define headers
    headers := []string{"ID", "Name", "Email", "Age", "Active", "Created At"}

    // Export to Excel
    err = excelManager.ExportToExcel(users, "Users", headers)
    if err != nil {
        http.Error(w, "Failed to export Excel", http.StatusInternalServerError)
        return
    }

    // Set response headers
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", "attachment; filename=users.xlsx")

    // Write to response
    err = excelManager.WriteToWriter(w)
    if err != nil {
        http.Error(w, "Failed to write Excel", http.StatusInternalServerError)
        return
    }
}
```

### Import Handler

```go
func ImportUsersHandler(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form
    err := r.ParseMultipartForm(10 << 20) // 10 MB max
    if err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    // Get uploaded file
    file, _, err := r.FormFile("excel_file")
    if err != nil {
        http.Error(w, "No file uploaded", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create Excel manager from file
    excelManager, err := excel.NewExcelManagerFromFile(fileHeader)
    if err != nil {
        http.Error(w, "Invalid Excel file", http.StatusBadRequest)
        return
    }

    // Import data
    data, err := excelManager.ImportFromExcel("Users", reflect.TypeOf(User{}))
    if err != nil {
        http.Error(w, "Failed to import Excel", http.StatusInternalServerError)
        return
    }

    users := data.([]User)

    // Validate data
    validators := map[string]func(string) error{
        "Name": func(value string) error {
            if value == "" {
                return fmt.Errorf("name is required")
            }
            return nil
        },
        "Email": func(value string) error {
            if !isValidEmail(value) {
                return fmt.Errorf("invalid email format")
            }
            return nil
        },
    }

    errors := excelManager.ValidateData(users, validators)
    if len(errors) > 0 {
        // Return validation errors
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "errors":  errors,
        })
        return
    }

    // Save users to database
    err = userService.BulkCreate(users)
    if err != nil {
        http.Error(w, "Failed to save users", http.StatusInternalServerError)
        return
    }

    // Return success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("Successfully imported %d users", len(users)),
    })
}
```

## Supported Data Types

| Go Type                                       | Excel/CSV Format | Notes                   |
| --------------------------------------------- | ---------------- | ----------------------- |
| `string`                                      | String           | Direct mapping          |
| `int`, `int8`, `int16`, `int32`, `int64`      | Number           | Automatic conversion    |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | Number           | Automatic conversion    |
| `float32`, `float64`                          | Number           | Automatic conversion    |
| `bool`                                        | Boolean          | true/false              |
| `time.Time`                                   | Date/Time        | Multiple format support |

## Time Format Support

The package supports multiple time formats for parsing:

- `2006-01-02 15:04:05`
- `2006-01-02`
- `01/02/2006`
- `2006-01-02T15:04:05Z`

## Validation Error Structure

```go
type ValidationError struct {
    Row    int    `json:"row"`    // Excel row number (1-based)
    Column string `json:"column"` // Column name
    Value  string `json:"value"`  // Invalid value
    Error  string `json:"error"`  // Error message
}
```

## Best Practices

1. **Use Struct Tags**: Always use `excel` or `json` tags for field mapping
2. **Validate Data**: Always validate imported data before saving
3. **Handle Errors**: Properly handle validation errors and provide feedback
4. **File Size Limits**: Set appropriate file size limits for uploads
5. **Memory Management**: For large files, consider streaming processing
6. **Error Messages**: Provide clear, user-friendly error messages
7. **Headers**: Always specify headers for better Excel formatting

## Performance Considerations

- **Large Files**: For files with >10,000 rows, consider streaming processing
- **Memory Usage**: Excel files are loaded entirely into memory
- **Validation**: Validation is performed in memory, consider batching for large datasets
- **Concurrent Processing**: The package is not thread-safe, use with caution in concurrent environments

## Example Integration with Chi Router

```go
func setupExcelRoutes(r *chi.Router) {
    r.Get("/export/users", ExportUsersHandler)
    r.Post("/import/users", ImportUsersHandler)
    r.Get("/export/users/csv", ExportUsersCSVHandler)
    r.Post("/import/users/csv", ImportUsersCSVHandler)
}
```
