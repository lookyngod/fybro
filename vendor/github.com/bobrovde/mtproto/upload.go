package mtproto

func (m *MTProto) UploadGetFile(location TL,limit int32) (*TL, error){
	return m.InvokeSync(TL_upload_getFile{Location:location,Limit:limit})
}

func (m *MTProto) UploadSaveFilePart(fileId int64, filePart int32, bytes []byte)  (*TL, error){
	return m.InvokeSync(TL_upload_saveFilePart{
		File_id:fileId,
		File_part:filePart,
		Bytes:bytes,
	})
}

func (m *MTProto) UploadSaveBigFilePart(fileId int64, filePart int32,fileTotalParts int32, bytes []byte)  (*TL, error){
	return m.InvokeSync(TL_upload_saveBigFilePart{
		File_id:fileId,
		File_part:filePart,
		File_total_parts:fileTotalParts,
		Bytes:bytes,
	})
}