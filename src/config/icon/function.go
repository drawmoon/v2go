package icon

func InitIcon(nerdfont bool) {
	if !nerdfont {
		Space = ""

		// file operations
		CompressFile = ""
		ExtractFile = ""
		Copy = ""
		Cut = ""
		Delete = ""

		// other
		Cursor = ">"
		Browser = ""
		Select = ""
		Error = ""
		Warn = ""
		Done = ""
		InOperation = ""
		Directory = ""
		Search = ""
		SortAsc = ""
		SortDesc = ""
	}
}
