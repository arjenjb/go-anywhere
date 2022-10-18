//go:build windows

package api

//func Test_GetIconFromResource(t *testing.T) {
//	resource, err := GetIconFromResource("D4480A50-BA28-11d1-8E75-00C04FA31A86", 0)
//	if err != nil {
//		return
//	}
//	print(resource)
//}

//func Test_ShortcutIconTarget(t *testing.T) {
//	target, index, err := ShortcutIconTarget("test\\custom-icon.lnk")
//	if err != nil {
//		return
//	}
//
//	ext := filepath.Ext(target)
//	if ext == ".dll" || ext == ".cpl" {
//		// If the icon is in a DLL we have to extract it
//		image, err := GetIconFromResource(target, index)
//
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		WriteImageToFile(image, "outimage.png")
//	}
//
//	log.Printf("Found target: %s", target)
//}
//
//func Test_AnotherShortcutIcon(t *testing.T) {
//	fname := "C:\\ProgramData\\Microsoft\\Windows\\Start Menu\\Programs\\Accessibility\\Speech Recognition.lnk"
//	target, index, err := ShortcutIconTarget(fname)
//	if err != nil {
//		t.Fatal(err)
//	}
//	ext := filepath.Ext(target)
//	if ext == ".dll" || ext == ".cpl" {
//		// If the icon is in a DLL we have to extract it
//		image, err := GetIconFromResource(target, index)
//
//		if err != nil {
//			t.Fatal(err)
//		}
//		WriteImageToFile(image, "outimage.png")
//	}
//}
//func Test_AnotherShortcutIcon2(t *testing.T) {
//	fname := "C:\\ProgramData\\Microsoft\\Windows\\Start Menu\\Programs\\Accessories\\Math Input Panel.lnk"
//	target, index, err := ShortcutIconTarget(fname)
//	if err != nil {
//		t.Fatal(err)
//	}
//	ext := filepath.Ext(target)
//	if ext == ".dll" || ext == ".cpl" {
//		// If the icon is in a DLL we have to extract it
//		image, err := GetIconFromResource(target, index)
//
//		if err != nil {
//			t.Fatal(err)
//		}
//		WriteImageToFile(image, "outimage.png")
//	}
//}
