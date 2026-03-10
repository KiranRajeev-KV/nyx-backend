package embedding

import (
	"fmt"
	"os"
	"path/filepath"

	ort "github.com/yalue/onnxruntime_go"
)

var globalService *EmbeddingService

type EmbeddingService struct {
	modelPath    string
	imagePreproc *ImagePreprocessor
	session      *ort.AdvancedSession
	inputNames   []string
	outputNames  []string
}

func NewEmbeddingService(modelPath string) (*EmbeddingService, error) {
	modelFile := filepath.Join(modelPath, "vision_model.onnx")
	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("ONNX model not found at %s", modelFile)
	}

	ort.SetSharedLibraryPath("/usr/local/lib/libonnxruntime.so.1.18.0")
	err := ort.InitializeEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}

	inputNames := []string{"pixel_values"}
	outputNames := []string{"embedding"}

	inputTensor, err := ort.NewTensor(ort.NewShape(1, 3, 224, 224), make([]float32, 1*3*224*224))
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	outputTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 512))
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	session, err := ort.NewAdvancedSession(
		modelFile,
		inputNames,
		outputNames,
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ONNX session: %w", err)
	}

	svc := &EmbeddingService{
		modelPath:    modelPath,
		imagePreproc: NewImagePreprocessor(),
		session:      session,
		inputNames:   inputNames,
		outputNames:  outputNames,
	}

	return svc, nil
}

func SetGlobalService(svc *EmbeddingService) {
	globalService = svc
}

func GetGlobalService() *EmbeddingService {
	return globalService
}

func (s *EmbeddingService) GetImageEmbedding(imageURL string) ([]float64, error) {
	inputData, err := s.imagePreproc.PreprocessFromURL(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess image: %w", err)
	}

	output, err := s.runInference(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}

	result := make([]float64, len(output))
	for i, v := range output {
		result[i] = float64(v)
	}

	return result, nil
}

func (s *EmbeddingService) runInference(inputData []float32) ([]float32, error) {
	inputTensor, err := ort.NewTensor(ort.NewShape(1, 3, 224, 224), inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	outputTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 512))
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	session, err := ort.NewAdvancedSession(
		filepath.Join(s.modelPath, "vision_model.onnx"),
		s.inputNames,
		s.outputNames,
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Destroy()

	err = session.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}

	return outputTensor.GetData(), nil
}

func IsServiceAvailable() bool {
	return globalService != nil
}
