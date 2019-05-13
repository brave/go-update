package extensiontest

import (
	"fmt"
)

// ExtensionRequestFnForXML creates a function for the specified appID which creates a function
// which takes in a version and returns an XML request.
func ExtensionRequestFnForXML(appID string) func(string) string {
	return func(version string) string {
		return fmt.Sprintf(`
		<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
			<hw physmemory="16"/>
			<os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
			<app appid="%s" version="%s" installsource="ondemand">
				<updatecheck />
				<ping rd="-2" ping_freshness="" />
		  </app>
		</request>`, appID, version)
	}
}

// ExtensionRequestFnForJSON creates a function for the specified appID which creates a function
// which takes in a version and returns an XML request.
func ExtensionRequestFnForJSON(appID string) func(string) string {
	return func(version string) string {
		return fmt.Sprintf(`{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"},"app":[{"appid":"%s","installsource":"ondemand","ping":{"r":-2},"updatecheck":{},"version":"%s"}]}}
`, appID, version)
	}
}

// ExtensionRequestFnForTwoXML creates a function for the specified appIDs which creates a function
// which takes in the appID versions and returns an XML request.
func ExtensionRequestFnForTwoXML(appID1 string, appID2 string) func(string, string) string {
	return func(version1 string, version2 string) string {
		return fmt.Sprintf(`
		<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.0" version="chrome-53.0.2785.116" prodversion="53.0.2785.116" requestid="{b4f77b70-af29-462b-a637-8a3e4be5ecd9}" lang="" updaterchannel="stable" prodchannel="stable" os="mac" arch="x64" nacl_arch="x86-64">
			<hw physmemory="16"/>
			<os platform="Mac OS X" version="10.11.6" arch="x86_64"/>
			<app appid="%s" version="%s" installsource="ondemand">
				<updatecheck />
				<ping rd="-2" ping_freshness="" />
		</app>
		<app appid="%s" version="%s" installsource="ondemand">
			<updatecheck />
			<ping rd="-2" ping_freshness="" />
		</app>
		</request>`, appID1, version1, appID2, version2)
	}
}

// ExtensionRequestFnForTwoJSON creates a function for the specified appIDs which creates a function
// which takes in the appID versions and returns an XML request.
func ExtensionRequestFnForTwoJSON(appID1 string, appID2 string) func(string, string) string {
	return func(version1 string, version2 string) string {
		return fmt.Sprintf(`
		{"request":{"protocol":"3.1","version":"chrome-53.0.2785.116","prodversion":"53.0.2785.116","requestid":"{e821bacd-8dbf-4cc8-9e8c-bcbe8c1cfd3d}","lang":"","updaterchannel":"stable","prodchannel":"stable","os":"mac","arch":"x64","nacl_arch":"x86-64","hw":{"physmemory":16},"os":{"arch":"x86_64","platform":"Mac OS X","version":"10.14.3"},"app":[{"appid":"%s","installsource":"ondemand","ping":{"r":-2},"updatecheck":{},"version":"%s"},{"appid":"%s","installsource":"ondemand","ping":{"r":-2},"updatecheck":{},"version":"%s"}]}}`, appID1, version1, appID2, version2)
	}
}
