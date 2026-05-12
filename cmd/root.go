package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wow-look-at-my/xml-validator/validator"
)

var schemaFile string

var rootCmd = &cobra.Command{
	Use:   "xml-validator [file]",
	Short: "Strict XML 1.1 validator",
	Long: `Validates XML documents strictly against the XML 1.1 specification.

Use --schema to also validate against an XSD schema:
  xml-validator --schema schema.xsd input.xml

Supported features:
  - Elements, attributes, text content, CDATA sections
  - Comments and processing instructions
  - Character references and predefined entity references
  - Namespace validation (Namespaces in XML 1.1)
  - UTF-8 and UTF-16 encodings with BOM detection
  - XML 1.1 line ending normalization
  - XSD schema validation (--schema)

Anything unsupported is a hard error:
  - DOCTYPE declarations
  - General entity references (beyond the 5 predefined)
  - XML 1.0 documents (version must be "1.1")
  - Missing XML declaration
  - Encodings other than UTF-8 and UTF-16`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if schemaFile != "" {
			return runWithSchema(args)
		}
		return runWellFormedness(args)
	},
}

func init() {
	rootCmd.Flags().StringVar(&schemaFile, "schema", "", "XSD schema file to validate against")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runWellFormedness(args []string) error {
	var input *os.File
	if len(args) == 0 {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("cannot open file: %w", err)
		}
		defer f.Close()
		input = f
	}

	if err := validator.Validate(input); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("valid XML 1.1 document")
	return nil
}

func runWithSchema(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("--schema requires an XML file argument (stdin not supported with schema validation)")
	}

	if err := validator.ValidateWithSchemaFile(args[0], schemaFile); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("valid XML 1.1 document (schema validated)")
	return nil
}
