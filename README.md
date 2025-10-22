# go-resume-builder

A simple command-line tool written in Go that generates professional PDF resumes from YAML or JSON input files.

## Description

This tool allows you to create clean, professional resumes by defining your information in a structured YAML or JSON format. It generates a PDF output using the gofpdf library, making it easy to maintain and update your resume programmatically.

## Installation

### Prerequisites
- Go 1.21 or later

### Install Dependencies
```bash
go mod download
```

### Build
```bash
go build -o resume-builder main.go
```

## How to Run

### Basic Usage
```bash
./resume-builder -input example-resume.yaml -output resume.pdf
```

### Command Line Options
- `-input`: Path to your resume data file (YAML or JSON format)
- `-output`: Path for the generated PDF file

### Example
```bash
# Using the provided example file
./resume-builder -input example-resume.yaml -output my-resume.pdf

# Using a JSON file
./resume-builder -input resume-data.json -output resume.pdf
```

## Input File Format

The tool accepts both YAML and JSON formats. See `example-resume.yaml` for a complete example of the expected structure including personal information, summary, experience, education, and skills sections.