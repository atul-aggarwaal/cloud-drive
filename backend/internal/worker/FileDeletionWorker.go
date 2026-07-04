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
	filesToDelete, err := worker.fileRepo.GetFilesMarkedForDeletion(ctx)

	if err != nil {
		return fmt.Errorf("loading files marked for deletion: %w", err)
	}

	for _, file := range filesToDelete {
		if err := worker.processFile(ctx, file); err != nil {
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

	// Mark requested file Deleted.
	if err := worker.fileRepo.MarkFileDeleted(ctx, file.ID); err != nil {
		return fmt.Errorf("deleting file metadata: %w", err)
	}
	return nil
}

func (worker *FileDeletionWorker) processFileVersions(ctx context.Context, file *domain.File) error {
	versions, err := worker.fileRepo.GetVersions(ctx, file.ID)
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
	for _, version := range versions {

		//First delete s3 version
		objectKey := fmt.Sprintf("user/%s/%s/%s", file.OwnerID, file.ID, version.VersionNum)
		err := worker.storage.DeleteObject(ctx, objectKey)
		if err != nil {
			log.Printf("delete s3 file version %s. error: %v:", objectKey, err)
			allDeleted = false
			failureCount += 1
			continue
		}

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
