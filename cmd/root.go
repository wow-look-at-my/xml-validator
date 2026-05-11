package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wow-look-at-my/xml-validator/validator"
)

var rootCmd = &cobra.Command{
	Use:   "xml-validator [file]",
	Short: "Strict XML 1.1 validator",
	Long: `Validates XML documents strictly against the XML 1.1 specification.

Supported features:
  - Elements, attributes, text content, CDATA sections
  - Comments and processing instructions
  - Character references (&#N; and &#xN;)
  - Predefined entity references (&amp; &lt; &gt; &apos; &quot;)
  - Namespace validation (Namespaces in XML 1.1)
  - UTF-8 and UTF-16 encodings
  - XML 1.1 line ending normalization (#x85, #x2028)
  - XML 1.1 character and name character classes

Anything unsupported is a hard error:
  - DOCTYPE declarations
  - General entity references (beyond the 5 predefined)
  - XML 1.0 documents (version must be "1.1")
  - Missing XML declaration
  - Encodings other than UTF-8 and UTF-16`,
	Args:    cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
