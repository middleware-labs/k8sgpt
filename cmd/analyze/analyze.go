package analyze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	explain   bool
	backend   string
	output    string
	filters   []string
	language  string
	nocache   bool
	namespace string
)

// AnalyzeCmd represents the problems command
var AnalyzeCmd = &cobra.Command{
	Use:     "analyze",
	Aliases: []string{"analyse"},
	Short:   "This command will find problems within your Kubernetes cluster",
	Long: `This command will find problems within your Kubernetes cluster and
	provide you with a list of issues that need to be resolved`,
	Run: func(cmd *cobra.Command, args []string) {

		// get backend from file
		backendType := viper.GetString("backend_type")
		if backendType == "" {
			color.Red("No backend set. Please run k8sgpt auth")
			os.Exit(1)
		}
		// override the default backend if a flag is provided
		if backend != "" {
			backendType = backend
		}
		// get the token with viper
		token := viper.GetString(fmt.Sprintf("%s_key", backendType))
		// check if nil
		if token == "" {
			color.Red("No %s key set. Please run k8sgpt auth", backendType)
			os.Exit(1)
		}

		var aiClient ai.IAI
		switch backendType {
		case "openai":
			aiClient = &ai.OpenAIClient{}
			if err := aiClient.Configure(token, language); err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
		default:
			color.Red("Backend not supported")
			os.Exit(1)
		}

		ctx := context.Background()
		// Get kubernetes client from viper
		client := viper.Get("kubernetesClient").(*kubernetes.Client)
		// Analysis configuration
		config := &analyzer.AnalysisConfiguration{
			Namespace: namespace,
			NoCache:   nocache,
			Explain:   explain,
		}

		var analysisResults *[]analyzer.Analysis = &[]analyzer.Analysis{}
		if err := analyzer.RunAnalysis(ctx, filters, config, client,
			aiClient, analysisResults); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		if len(*analysisResults) == 0 {
			color.Green("{ \"status\": \"OK\" }")
			os.Exit(0)
		}
		var bar = progressbar.Default(int64(len(*analysisResults)))
		if !explain {
			bar.Clear()
		}
		var printOutput []analyzer.Analysis

		for _, analysis := range *analysisResults {

			if explain {
				parsedText, err := analyzer.ParseViaAI(ctx, config, aiClient, analysis.Error)
				if err != nil {
					// Check for exhaustion
					if strings.Contains(err.Error(), "status code: 429") {
						color.Red("Exhausted API quota. Please try again later")
						os.Exit(1)
					}
					color.Red("Error: %v", err)
					continue
				}
				analysis.Details = parsedText
				bar.Add(1)
			}
			printOutput = append(printOutput, analysis)
		}

		// print results
		for _, analysis := range printOutput {

			switch output {
			case "json":
				analysis.Error = analysis.Error[0:]
				_, err := json.Marshal(analysis)
				if err != nil {
					color.Red("Error: %v", err)
					os.Exit(1)
				}
				// fmt.Println(string(j))
			default:
				// fmt.Printf("%s %s(%s)\n", color.CyanString("%d", n),
				// color.YellowString(analysis.Name), color.CyanString(analysis.ParentObject))
				for _, err := range analysis.Error {
					fmt.Printf("- %s %s\n", color.RedString("Error:"), color.RedString(err))
				}
				fmt.Println(color.GreenString(analysis.Details + "\n"))
			}
		}
	},
}

func GptAnalysis() {

	var aiClient ai.IAI
	switch "openai" {
	case "openai":
		aiClient = &ai.OpenAIClient{}
		// if err := aiClient.Configure("", language); err != nil {
		// 	color.Red("Error: %v", err)
		// 	os.Exit(1)
		// }
	default:
		color.Red("Backend not supported")
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := kubernetes.NewClient("", "")
	if err != nil {
		fmt.Println("Error creating client:", err)
	}
	// Analysis configuration
	config := &analyzer.AnalysisConfiguration{
		Namespace: namespace,
		NoCache:   nocache,
		Explain:   true,
	}

	var analysisResults *[]analyzer.Analysis = &[]analyzer.Analysis{}
	if err := analyzer.RunAnalysis(ctx, filters, config, client,
		aiClient, analysisResults); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// var bar *progressbar.ProgressBar
	if len(*analysisResults) > 0 {
		// bar = progressbar.Default(int64(len(*analysisResults)))
	} else {
		color.Green("{ \"status\": \"OK\" }")
		os.Exit(0)
	}

	// This variable is used to store the results that will be printed
	// It's necessary because the heap memory is lost when the function returns
	var printOutput []analyzer.Analysis

	for _, analysis := range *analysisResults {

		// if explain {
		// 	parsedText, err := analyzer.ParseViaAI(ctx, config, aiClient, analysis.Error)
		// 	if err != nil {
		// 		// Check for exhaustion
		// 		if strings.Contains(err.Error(), "status code: 429") {
		// 			color.Red("Exhausted API quota. Please try again later")
		// 			os.Exit(1)
		// 		}
		// 		color.Red("Error: %v", err)
		// 		continue
		// 	}
		// 	analysis.Details = parsedText
		// 	bar.Add(1)
		// }
		printOutput = append(printOutput, analysis)
	}

	// print results
	for _, analysis := range printOutput {

		switch output {
		case "json":
			analysis.Error = analysis.Error[0:]
			_, err := json.Marshal(analysis)
			if err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
			// fmt.Println(string(j))
		default:

			// fmt.Printf("%s %s(%s)\n", color.CyanString("%d", n),
			// color.YellowString(analysis.Name), color.CyanString(analysis.ParentObject))
			for _, err := range analysis.Error {
				sendErrorsToMiddleware(err, analysis.Name, analysis.ParentObject)
				// fmt.Printf("- %s %s\n", color.RedString("Error:"), color.RedString(err))
			}
			color.GreenString(analysis.Details)
		}
	}
}

func sendErrorsToMiddleware(message string, name string, parent string) {

	// fmt.Println("sendErrorsToMiddleware--------------------------------")
	apiKey := ""
	target := ""

	if val, ok := os.LookupEnv("MW_API_KEY"); ok {
		apiKey = val
	}

	if val2, ok2 := os.LookupEnv("TARGET"); ok2 {
		target = val2
	}

	body := []byte(`{
		"resource_logs": [
		  {
			"resource": {
			  "attributes": [
				{
				  "key": "mw.account_key",
				  "value": {
					"string_value": "` + apiKey + `"
				  }
				},
				{
				  "key": "mw.resource_type",
				  "value": {
					"string_value": "custom"
				  }
				}
			  ]
			},
			"scope_logs": [
				{
				  "log_records": [
					  {
						  "attributes": [
							{
							  "key": "device",
							  "value": {
								"string_value": "nvme0n1p4"
							  }
							}
						  ],
						  "body": {
							  "string_value": "` + message + `"
							 
						  },
						  "severity_number": 17,
						  "severity_text": "ERROR",
						  "time_unix_nano": ` + strconv.FormatInt(time.Now().UnixNano(), 10) + `,
						  "observed_time_unix_nano": ` + strconv.FormatInt(time.Now().UnixNano(), 10) + `
					  }
				  ]   
				}
			 ]
		  }
		]
	  } 
	  `)

	request, err := http.NewRequest("POST", target+"/v1/logs", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Could not send data to Middleware", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("authorization", apiKey)
	// fmt.Println("http request created")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
}

func init() {

	// namespace flag
	AnalyzeCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to analyze")
	// no cache flag
	AnalyzeCmd.Flags().BoolVarP(&nocache, "no-cache", "c", false, "Do not use cached data")
	// array of strings flag
	AnalyzeCmd.Flags().StringSliceVarP(&filters, "filter", "f", []string{}, "Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)")
	// explain flag
	AnalyzeCmd.Flags().BoolVarP(&explain, "explain", "e", false, "Explain the problem to me")
	// add flag for backend
	AnalyzeCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
	// output as json
	AnalyzeCmd.Flags().StringVarP(&output, "output", "o", "text", "Output format (text, json)")
	// add language options for output
	AnalyzeCmd.Flags().StringVarP(&language, "language", "l", "english", "Languages to use for AI (e.g. 'English', 'Spanish', 'French', 'German', 'Italian', 'Portuguese', 'Dutch', 'Russian', 'Chinese', 'Japanese', 'Korean')")
}
