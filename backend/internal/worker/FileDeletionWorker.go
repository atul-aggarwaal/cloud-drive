package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
)
//A periodic worker which will periodically scan files marked as deleted and delete them from s3 bucket along with its file versions metadata
// if everything is successful, it will mark file as "DELETED"
type FileDeletionWorker struct {
	fileRepo domain.FileRepository
	storage  domain.BlobStorage
}

func NewFileDeletionWorker(fileRepo domain.FileRepository, storage domain.BlobStorage) *FileDeletionWorker {
	return &FileDeletionWorker{
		fileRepo: fileRepo,
		storage:  storage,
	}
}
func (worker *FileDeletionWorker) RunOnce(ctx context.Context) error {
	// TODO: No concurrency guard. GetFilesMarkedForDeletion selects on
	// lifecycle_status = DELETE_REQUESTED with no claiming, so two worker
	// instances (or overlapping ticks) can process the same file. Idempotency
	// keeps state correct, but we double work. Claim rows (SELECT ... FOR UPDATE
	// SKIP LOCKED, or flip to an intermediate DELETING status) before processing.
	filesToDelete, err := worker.fileRepo.GetFilesMarkedForDeletion(ctx)

	log.Printf("files found for deletion: %d", len(filesToDelete))
	if err != nil {
		return fmt.Errorf("loading files marked for deletion: %w", err)
	}

	for _, file := range filesToDelete {
		log.Printf("processing file: %s", file.FileName)
		if err := worker.processFile(ctx, file); err != nil {
			// TODO: No retry ceiling / dead-letter. A version whose S3 delete
			// genuinely errors (e.g. IAM change) keeps this file in
			// DELETE_REQUESTED forever, reprocessed every cycle with no alerting.
			// Add an attempt counter and a terminal FAILED status.
			log.Printf(
				"failed processing file id=%s name=%s: %v",
				file.ID,
				file.FileName,
				err,
			)
			continue
		}

	}

	return nil
}

func (worker *FileDeletionWorker) processFile(ctx context.Context, file *domain.File) error {

	//delete all Versions related tof file
	err := worker.processFileVersions(ctx, file)
	if err != nil {
		return fmt.Errorf("deleting versions for file %s: %w", file.ID, err)
	}

	log.Println("marking file as deleted")
	// Mark requested file Deleted.
	if err := worker.fileRepo.MarkFileDeleted(ctx, file.ID); err != nil {
		return fmt.Errorf("deleting file metadata: %w", err)
	}
	return nil
}

func (worker *FileDeletionWorker) processFileVersions(ctx context.Context, file *domain.File) error {
	log.Printf("processing file versions for file: %s", file.FileName)
	versions, err := worker.fileRepo.GetVersions(ctx, file.ID)
	log.Printf("total versions found: %d", len(versions))
	if err != nil {
		return fmt.Errorf(
			"load versions for file %s: %w",
			file.ID,
			err,
		)
	}

	allDeleted := true
	failureCount := 0
	//Delete each version one by one
	log.Println("processing versions for deletion")
	for _, version := range versions {

		//First delete s3 version
		objectKey := fmt.Sprintf("user/%s/%s/v%d", file.OwnerID, file.ID, version.VersionNum)
		log.Printf("deleting version key %s from s3", objectKey)
		err := worker.storage.DeleteObject(ctx, objectKey)
		if err != nil {
			log.Printf("delete s3 file version %s. error: %v:", objectKey, err)
			allDeleted = false
			failureCount += 1
			continue
		}

		log.Println("deleting version from db.")
		// if version successfully deleted from s3, delete its metadata
		err = worker.fileRepo.DeleteFileVersion(ctx, version.VersionNum, version.FileId)
		if err != nil { //confinue processing other versions, if one failed
			log.Printf("delete file version %s. error: %v:", version.ID, err)
			failureCount += 1
			allDeleted = false
			continue
		}
	}

	if !allDeleted {
		return fmt.Errorf("failed to delete %d versions: ", failureCount)
	}
	return nil
}
