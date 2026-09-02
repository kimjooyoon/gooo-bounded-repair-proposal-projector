package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-bounded-repair-proposal-projector/internal/projector"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "project":
		return project(args[1:], stdout, stderr)
	case "conformance":
		return conformance(args[1:], stdout, stderr)
	case "compile":
		return compile(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, "gooo-bounded-repair-proposal-projector/v0.1.0")
		return 0
	default:
		usage(stderr)
		return 2
	}
}

func usage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo-repair-proposal-projector <project|conformance|compile|version>")
}

func project(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("project", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "input repository root")
	sourcePath := flags.String("source", ".gooo/repair-proposal-projector.gooo", "authoritative .gooo source")
	inputPath := flags.String("input", "", "immutable graph and evidence input")
	outputDir := flags.String("output-dir", "", "caller-owned absolute output directory")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *inputPath == "" || !filepath.IsAbs(*outputDir) {
		fmt.Fprintln(stderr, "project requires -input and an absolute -output-dir")
		return 2
	}
	if err := projector.EnsureCallerOwnedOutput(*outputDir, *root); err != nil {
		fmt.Fprintf(stderr, "project output: %v\n", err)
		return 1
	}
	evaluation, err := loadEvaluation(*root, *sourcePath, *inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "project: %v\n", err)
		return 1
	}
	if err := projector.WriteEvaluation(*outputDir, evaluation); err != nil {
		fmt.Fprintf(stderr, "project write: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s state=%s proposal=%t output=%s\n", evaluation.Dossier.CaseID, evaluation.Dossier.State, evaluation.Dossier.Proposal != nil, *outputDir)
	return 0
}

func compile(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("compile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sourcePath := flags.String("source", ".gooo/repair-proposal-projector.gooo", "authoritative .gooo source")
	outputIR := flags.String("output-ir", "", "caller-owned absolute semantic IR path")
	outputEvaluator := flags.String("output-evaluator", "", "caller-owned absolute generated evaluator path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if !filepath.IsAbs(*outputIR) || !filepath.IsAbs(*outputEvaluator) {
		fmt.Fprintln(stderr, "compile requires absolute -output-ir and -output-evaluator paths")
		return 2
	}
	sourceRaw, err := os.ReadFile(*sourcePath)
	if err != nil {
		fmt.Fprintf(stderr, "compile source: %v\n", err)
		return 1
	}
	ir, _, err := projector.CompileSource(*sourcePath, sourceRaw)
	if err != nil {
		fmt.Fprintf(stderr, "compile semantic source: %v\n", err)
		return 1
	}
	irRaw, err := json.MarshalIndent(ir, "", "  ")
	if err != nil {
		return 1
	}
	irRaw = append(irRaw, '\n')
	generated := projector.GenerateEvaluator(ir, projector.DigestBytes(irRaw), projector.DigestBytes(sourceRaw))
	if err := os.WriteFile(*outputIR, irRaw, 0o644); err != nil {
		return 1
	}
	if err := os.WriteFile(*outputEvaluator, generated, 0o644); err != nil {
		return 1
	}
	fmt.Fprintf(stdout, "semantic_ir=%s generated_evaluator=%s\n", *outputIR, *outputEvaluator)
	return 0
}

func conformance(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "input repository root")
	sourcePath := flags.String("source", ".gooo/repair-proposal-projector.gooo", "authoritative .gooo source")
	corpusPath := flags.String("corpus", "fixtures/corpus.json", "fixed conformance corpus")
	outputDir := flags.String("output-dir", "", "caller-owned absolute output directory")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if !filepath.IsAbs(*outputDir) {
		fmt.Fprintln(stderr, "conformance requires an absolute -output-dir")
		return 2
	}
	index, err := projector.RunConformance(*root, *sourcePath, *corpusPath, *outputDir)
	if err != nil {
		fmt.Fprintf(stderr, "conformance: %v\n", err)
		return 1
	}
	for _, result := range index.Results {
		fmt.Fprintf(stdout, "%s class=%s state=%s incidents=%d\n", result.CaseID, result.Class, result.State, len(result.IncidentIDs))
	}
	return 0
}

func loadEvaluation(root, sourcePath, inputPath string) (projector.Evaluation, error) {
	sourceRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		return projector.Evaluation{}, err
	}
	ir, _, err := projector.CompileSource(sourcePath, sourceRaw)
	if err != nil {
		return projector.Evaluation{}, err
	}
	input, inputRaw, err := projector.ReadInput(inputPath)
	if err != nil {
		return projector.Evaluation{}, err
	}
	dossier, events, err := projector.Evaluate(input, projector.DigestBytes(inputRaw), ir)
	if err != nil {
		return projector.Evaluation{}, err
	}
	irRaw, err := json.MarshalIndent(ir, "", "  ")
	if err != nil {
		return projector.Evaluation{}, err
	}
	irRaw = append(irRaw, '\n')
	return projector.Evaluation{Dossier: dossier, Events: events, SemanticIRRaw: irRaw, GeneratedRaw: projector.GenerateEvaluator(ir, projector.DigestBytes(irRaw), projector.DigestBytes(sourceRaw)), Replay: projector.ReplayReceipt{}}, nil
}
