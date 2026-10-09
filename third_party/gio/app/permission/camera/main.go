// SPDX-License-Identifier: Unlicense OR MIT

/*
Package camera implements permissions to access camera hardware.

# Android

The following entries will be added to AndroidManifest.xml:

	<uses-permission android:name="android.permission.CAMERA"/>
	<uses-feature android:name="android.hardware.camera" android:required="false"/>

CAMERA is a "dangerous" permission. See documentation for package
github.com/dyike/keel/third_party/gio/app/permission for more information.
*/
package camera
