import QtQuick 2.13
import QtQuick.Controls 2.13
import QtQuick.Layouts 1.13
import QtQuick.Controls.Material 2.12
import QtQuick.Dialogs 1.3
import QtQuick.Shapes 1.13
import csc 1.2 as CSC

Page {
    id: page
    padding: 2 * CSC.Style.padding
    contentHeight: stack.height
    
    Material.accent: CSC.Style.primaryColor
    Material.foreground: CSC.Style.grey

    property bool chosen: false

    FileDialog {
        id: dialogSelect
        title: "Select file to export"
        folder: shortcuts.home
        selectExisting: true
        selectFolder: false

        onAccepted: {
            page.chosen = true
            exportModel.setProperty(0, "name", dialogSelect.fileUrl.toString())
        }
    }

    Connections {
        target: QmlBridge
        onExportFinished: if (success) {
            stack.currentIndex = 4
        } else {
            stack.currentIndex = 2
        }
    }

    ListModel {
        id: exportModel

        ListElement {
            name: ""
            bucket: ""
            modifiable: true
        }
    }

    header: CSC.ProgressTracker {
        id: tracker
        visible: stack.currentIndex >= 1
        progressIndex: stack.currentIndex - 1
        model: ["Choose directory", "Export files", "Export complete"]
    }

    StackLayout {
        id: stack
        width: parent.width
        height: children[currentIndex].height
        currentIndex: QmlBridge.isProjectManager ? 1 : 0

        ColumnLayout {
            spacing: CSC.Style.padding

            Label {
                text: "<h1>Export is not possible</h1>"
                maximumLineCount: 1
            }

            Label {
                text: "Your need to be project manager to export files."
                font.pixelSize: 14
            }
        }

        FocusScope {
            focus: visible 
            height: childrenRect.height
            width: childrenRect.width
            
            ColumnLayout {
                id: folderColumn
                spacing: CSC.Style.padding
                width: stack.width

                Keys.onReturnPressed: continueButton.clicked() // Enter key
                Keys.onEnterPressed: continueButton.clicked()  // Numpad enter key

                Label {
                    text: "<h1>Select a destination folder for your export</h1>"
                    maximumLineCount: 1
                }

                Label {
                    text: "Your export will be sent to SD Connect. If folder does not already exist in SD Connect, it will be created. Please note that the folder name cannot be modified afterwards."
                    wrapMode: Text.Wrap
                    lineHeight: 1.2
                    font.pixelSize: 14
                    Layout.maximumWidth: parent.width
                }

                CSC.TextField {
                    id: nameField
                    titleText: "Folder name"
                    implicitWidth: 400

                    property string compareText: ""
                    property real spaceLeft

                    onVisibleChanged: spaceLeft = Qt.binding(function() { return this.mapFromItem(null, 0, window.height).y - this.height })

                    onFocusChanged: if (focus) {
                        popupBuckets.open()
                    }

                    onTextChanged: {
                        if (focus) {
                            compareText = text
                        }
                        if (text != "") {
                            popupBuckets.open()
                        }
                    }

                    Popup {
                        id: popupBuckets
                        y: parent.height
                        width: parent.width
                        implicitHeight: Math.min(parent.spaceLeft - CSC.Style.padding, contentItem.implicitHeight)
                        padding: 0
                        closePolicy: Popup.NoAutoClose

                        Material.elevation: bucketsList.implicitHeight > 0 ? 6 : 0

                        contentItem: ListView {
                            id: bucketsList
                            clip: true
                            implicitHeight: contentHeight
                            currentIndex: 0
                            model: QmlBridge.buckets

                            delegate: ItemDelegate {
                                height: modelData.includes(nameField.compareText) ? Math.max(50, implicitHeight) : 0
                                width: nameField.width
                                highlighted: bucketsList.currentIndex === index

                                onClicked: {
                                    bucketsList.currentIndex = index
                                    nameField.focus = false
                                    folderColumn.focus = true
                                    nameField.text = modelData
                                    popupBuckets.close()
                                }

                                contentItem: Label {
                                    text: modelData
                                    color: CSC.Style.grey
                                    font.pixelSize: 15
                                    font.weight: Font.Medium 
                                    wrapMode: Text.Wrap
                                    visible: parent.height > 0
                                    verticalAlignment: Text.AlignVCenter
                                }

                                background: Rectangle {
                                    anchors.fill: parent
                                    color: (parent.hovered || parent.highlighted) ? CSC.Style.lightBlue : "transparent"
                                }
                            }

                            ScrollIndicator.vertical: ScrollIndicator { }
                        }
                    }
                }

                CSC.Button {
                    id: continueButton
                    text: "Continue"
                    enabled: nameField.text != ""
                    Layout.alignment: Qt.AlignRight

                    onClicked: if (enabled) { 
                        popupBuckets.close()
                        exportModel.setProperty(0, "bucket", nameField.text)
                        stack.currentIndex = stack.currentIndex + 1 
                    }
                }
            }
        }

        FocusScope {
            focus: visible
            height: childrenRect.height
            width: childrenRect.width

            ColumnLayout {
                spacing: CSC.Style.padding
                width: stack.width

                Keys.onReturnPressed: exportButton.clicked() // Enter key
                Keys.onEnterPressed: exportButton.clicked()  // Numpad enter key

                DropArea {
                    id: dropArea
                    visible: !page.chosen
                    Layout.preferredHeight: dragColumn.height
                    Layout.fillWidth: true

                    Shape {
                        id: shape
                        anchors.fill: parent

                        ShapePath {
                            fillColor: "transparent"
                            strokeWidth: 3
                            strokeColor: dropArea.containsDrag ? CSC.Style.primaryColor : Qt.rgba(CSC.Style.primaryColor.r, CSC.Style.primaryColor.g, CSC.Style.primaryColor.b, 0.5)
                            strokeStyle: ShapePath.DashLine
                            dashPattern: [ 1, 3 ]
                            startX: 0; startY: 0
                            PathLine { x: shape.width; y: 0 }
                            PathLine { x: shape.width; y: shape.height }
                            PathLine { x: 0; y: shape.height }
                            PathLine { x: 0 ; y: 0 }
                        }
                    }

                    Column {
                        id: dragColumn
                        padding: 40
                        spacing: CSC.Style.padding
                        anchors.horizontalCenter: parent.horizontalCenter

                        Row {
                            id: dragRow
                            spacing: CSC.Style.padding
                            anchors.horizontalCenter: parent.horizontalCenter

                            Label {
                                text: "Drag and drop file or"
                                font.pixelSize: 15
                                font.weight: Font.DemiBold
                                anchors.verticalCenter: selectButton.verticalCenter
                            }

                            CSC.Button {
                                id: selectButton
                                text: "Select file"
                                outlined: true

                                onClicked: dialogSelect.visible = true
                            }
                        }

                        Label {
                            text: "If you wish to export multiple files, please create a tar/zip file." 
                            font.pixelSize: 14
                            anchors.horizontalCenter: dragRow.horizontalCenter
                        }
                    }

                    onDropped: {
                        if (!drop.hasUrls) {
                            popup.errorMessage = "Dropped item was not a file"
                            popup.open()
                            return
                        }

                        if (drop.urls.length != 1) {
                            popup.errorMessage = "Dropped too many items"
                            popup.open()
                            return
                        }
                        
                        if (QmlBridge.isFile(drop.urls[0])) {
                            page.chosen = true
                            exportModel.setProperty(0, "name", drop.urls[0])
                        } else {
                            popup.errorMessage = "Dropped item was not a file"
                            popup.open()
                        }
                    }
                }

                CSC.Table {
                    visible: !dropArea.visible
                    hiddenHeader: true
                    objectName: "files"
                    Layout.fillWidth: true

                    contentSource: fileLine
                    footerSource: footerLine
                    modelSource: exportModel
                }

                RowLayout {
                    spacing: CSC.Style.padding

                    CSC.Button {
                        text: "Cancel"
                        outlined: true

                        onClicked: { 
                            stack.currentIndex = stack.currentIndex - 1
                            page.chosen = false
                        }
                    }

                    Item {
                        Layout.fillWidth: true
                    }

                    CSC.Button {
                        id: exportButton
                        text: "Export"
                        enabled: page.chosen

                        onClicked: if (enabled) { 
                            exportModel.setProperty(0, "modifiable", false)
                            stack.currentIndex = stack.currentIndex + 1 
                            QmlBridge.exportFile(exportModel.get(0).bucket, exportModel.get(0).name) 
                        }
                    }
                }
            }
        }

        ColumnLayout {
            spacing: CSC.Style.padding
            focus: visible

            Label {
                text: "<h1>Exporting file</h1>"
                maximumLineCount: 1
            }

            ColumnLayout {
                spacing: 0.5 * CSC.Style.padding

                Label {
                    text: "Please wait, this might take a few minutes."
                    font.pixelSize: 14
                    maximumLineCount: 1
                }

                CSC.ProgressBar {
                    indeterminate: true
                    Layout.fillWidth: true
                }
            }

            CSC.Table {
                hiddenHeader: true
                objectName: "files"
                Layout.fillWidth: true

                contentSource: fileLine
                footerSource: footerLine
                modelSource: exportModel
            }
        }

        FocusScope {
            focus: visible
            height: childrenRect.height
            width: childrenRect.width

            ColumnLayout {
                spacing: CSC.Style.padding
                width: stack.width

                Keys.onReturnPressed: newButton.clicked() // Enter key
                Keys.onEnterPressed: newButton.clicked()  // Numpad enter key

                Label {
                    text: "<h1>Export complete</h1>"
                    maximumLineCount: 1
                }

                Label {
                    text: "The file has been uploaded to SD Connect. You can now close or minimise the window to continue working."
                    font.pixelSize: 14
                }

                CSC.Button {
                    id: newButton
                    text: "New export"
                    Layout.alignment: Qt.AlignRight

                    onClicked: { nameField.text = ""; stack.currentIndex = 1 }
                }
            }
        }
    }

    Component {
        id: footerLine

        RowLayout {
            Label {
                text: "File Name"
                Layout.preferredWidth: parent.width * 0.4
            }

            Label {
                text: "Target Folder"
                Layout.fillWidth: true
            }
        }
    }

    Component {
        id: fileLine

        RowLayout {
            property string name: modelData ? modelData.name : ""
            property string bucket: modelData ? modelData.bucket : ""
            property bool modifiable: modelData ? modelData.modifiable : false

            Label {
                text: parent.name.split('/').reverse()[0]
                elide: Text.ElideRight
                Layout.preferredWidth: parent.width * 0.4
            }

            Label {
                text: parent.bucket
                elide: Text.ElideRight
                Layout.fillWidth: true
            }

            Button {
                id: remove
                text: "Remove"
                visible: parent.modifiable
                enabled: visible
                font.pixelSize: 14
                font.weight: Font.DemiBold
                icon.source: "qrc:/qml/images/delete.svg"

                Material.foreground: CSC.Style.primaryColor

                onClicked: page.chosen = false

                MouseArea {
                    cursorShape: Qt.PointingHandCursor
                    acceptedButtons: Qt.NoButton
                    anchors.fill: parent
                }

                background: Rectangle {
                    color: "transparent"
                }
            }
        }
    }
}