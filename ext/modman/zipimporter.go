package modman

//func ImportFromZip(fsys fs.FS, zipFile, fileName string) (tengo.Importable, error) {
//	fi, err := fs.Stat(fsys, zipFile)
//	if err != nil {
//		return nil, err
//	}
//
//	reader, err := fsys.Open(zipFile)
//	if err != nil {
//		return nil, err
//	}
//	rr, _ := reader.(io.ReaderAt)
//	zr, err := zip.NewReader(rr, fi.Size())
//	if err != nil {
//		return nil, err
//	}
//	defer func() {
//		err = reader.Close()
//		log.Info("close zip file:", zipFile, err)
//	}()
//	fileReader, err := zr.Open(fileName)
//	if err != nil {
//		return nil, err
//	}
//	defer func(fs fs.File) {
//		err = fs.Close()
//		if err != nil {
//			log.Error(err)
//		}
//	}(fileReader)
//	if fi, err := fileReader.Stat(); err != nil {
//		return nil, err
//	} else {
//		if fi.IsDir() {
//			return nil, fmt.Errorf("%s is a directory", fileName)
//		}
//	}
//	bytes, err := ioutil.ReadAll(fileReader)
//	if err != nil {
//		return nil, err
//	}
//	bytes, _ = transpile.Transpile(bytes)
//	return &tengo.SourceModule{Src: bytes}, nil
//}
//
//func ZipImporter(zipFile, fileName string) ImportFunc {
//	return func(s string) tengo.Importable {
//		reader, err := zip.OpenReader(zipFile)
//		if err != nil {
//			log.Error(err)
//			return nil
//		}
//		fileReader, err := reader.Open(fileName)
//		if err != nil {
//			log.Error(err)
//			return nil
//		}
//		defer func() {
//			err := fileReader.Close()
//			if err != nil {
//				log.Error(err)
//				return
//			}
//		}()
//		if fi, err := fileReader.Stat(); err != nil {
//			log.Error(err)
//			return nil
//		} else {
//			if fi.IsDir() {
//				log.Errorf("%s is a directory", fileName)
//				return nil
//			}
//		}
//		bytes, err := ioutil.ReadAll(fileReader)
//		if err != nil {
//			log.Error(err)
//			return nil
//		}
//		bytes, _ = transpile.Transpile(bytes)
//		return &tengo.SourceModule{Src: bytes}
//	}
//}
