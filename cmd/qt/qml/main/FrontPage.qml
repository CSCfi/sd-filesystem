import QtQuick 2.13
import QtQuick.Controls 2.13
import QtQuick.Layouts 1.13
import QtQuick.Controls.Material 2.12
import QtQuick.Dialogs 1.3
import csc 1.2 as CSC

Page {
    id: page
    padding: 2 * CSC.Style.padding

    FileDialog {
        id: dialogCreate
        title: "Choose or create a folder"
        folder: shortcuts.home
        selectExisting: false
        selectFolder: true
        onAccepted: {
            var mountError = QmlBridge.changeMountPoint(folder)
            if (mountError) {
                popup.errorMessage = mountError
                popup.open()
            }
        }
    }

    Connections {
		target: QmlBridge
        onPanic: {
            popupPanic.closePolicy = Popup.NoAutoClose // User must choose ignore or quit
            popupPanic.open()
        }
	}

    Connections {
        target: dialogSave
        onReady: {
            if (ignoreButton.checked) {
                popupPanic.close()
            } else if (quitButton.checked) {
                close()
            }
        }
    }

    CSC.Popup {
		id: popupPanic
		errorMessage: "How can this be! Data Gateway failed to load correctly.\nSave logs to find out why this happened and either quit the application or continue at your own peril..."
        
		ColumnLayout {
			width: parent.width

			CheckBox {
				id: logCheck
				checked: true
				text: "Yes, save logs to file"

				Material.accent: CSC.Style.primaryColor
			}

			Row {
				spacing: CSC.Style.padding
				Layout.alignment: Qt.AlignRight

				CSC.Button {
					id: ignoreButton
					text: "Ignore"
					outlined: true
					checkable: true
                    //mainColor: CSC.Style.red

					onClicked: {
						if (logCheck.checked) {
							dialogSave.visible = true
						} else {
							popupPanic.close()
						}
					}
				}

				CSC.Button {
					id: quitButton
					text: "Quit"
					checkable: true
                    //mainColor: CSC.Style.red
					
					onClicked: {
						if (logCheck.checked) {
							dialogSave.visible = true
						} else {
							close()
						}
					}
				}
			}
		}
	}

    header: CSC.ProgressTracker {
        id: tracker
        progressIndex: stack.currentIndex
    }

    contentItem: StackLayout {
        id: stack
        currentIndex: 0

        ColumnLayout {
            spacing: CSC.Style.padding
            focus: visible

            Keys.onReturnPressed: continueButton.clicked() // Enter key
            Keys.onEnterPressed: continueButton.clicked()  // Numpad enter key

            Text {
                text: "<h1>Choose directory</h1>"
                color: CSC.Style.grey
                maximumLineCount: 1
            }

            Text {
                text: "Choose in which local directory your data will be available"
                color: CSC.Style.grey
                font.pixelSize: 13
                maximumLineCount: 1
            }

            Row {
                spacing: CSC.Style.padding

                Rectangle {
                    radius: 5
                    border.width: 1
                    border.color: CSC.Style.grey
                    width: 350
                    height: childrenRect.height
                    anchors.verticalCenter: changeButton.verticalCenter

                    Flickable {
                        clip: true
                        width: parent.width
                        height: mountText.height
                        contentWidth: mountText.width
                        boundsBehavior: Flickable.StopAtBounds

                        ScrollBar.horizontal: ScrollBar { interactive: false }
                        
                        Text {
                            id: mountText
                            text: QmlBridge.mountPoint
                            font.pixelSize: 15
                            verticalAlignment: Text.AlignVCenter
                            maximumLineCount: 1
                            padding: 10
                        }
                    }
                }

                CSC.Button {
                    id: changeButton
                    text: "Change"
                    outlined: true

                    onClicked: { popup.close(); dialogCreate.visible = true }
                }
            }

            CSC.Button {
                id: continueButton
                text: "Continue"
                onClicked: { stack.currentIndex = 1; QmlBridge.loadFuse() }
            }
        }      

        ColumnLayout {
            id: accessLayout
            spacing: CSC.Style.padding
            focus: visible

            Keys.onReturnPressed: openButton.clicked() // Enter key
            Keys.onEnterPressed: openButton.clicked()  // Numpad enter key

            Text {
                id: headerText
                text: "<h1>Preparing access</h1>"
                color: CSC.Style.grey
                maximumLineCount: 1
            }

            RowLayout {
                id: buttonRow
                spacing: CSC.Style.padding
                visible: false

                CSC.Button {
                    id: openButton
                    text: "Open folder" 

                    onClicked: QmlBridge.openFuse()
                }

                CSC.Button {
                    id: refreshButton
                    text: "Refresh"
                    outlined: true

                    Material.accent: "white"

                    onClicked: {
                        state = "loading"
                        var message = QmlBridge.refreshFuse()
                        if (message != "") {
                            createButton.state = ""
                            popup.errorMessage = message
                            popup.open()
                        }
                    }

                    Connections {
                        target: QmlBridge
                        onFuseReady: refreshButton.state = ""
                    }

                    states: [
                        State {
                            name: "loading";  
                            PropertyChanges { target: refreshButton; enabled: false }
                            PropertyChanges { target: openButton; enabled: false }
                        }
                    ]
                }
            }

            ColumnLayout {
                spacing: 0.5 * CSC.Style.padding

                Text {
                    id: infoText
                    text: "Please wait, this might take a few minutes."
                    color: CSC.Style.grey
                    font.pixelSize: 13
                    maximumLineCount: 1
                }

                CSC.ProgressBar {
                    id: progressbar
                    value: ProjectModel.loadedContainers / ProjectModel.allContainers
                    Layout.fillWidth: true
                }

                Text {
                    text: Math.floor(progressbar.value * 100) + "% complete"
                    color: CSC.Style.grey
                    font.pixelSize: 13
                    maximumLineCount: 1
                }
            }

            CSC.Table {
                id: table
                modelSource: ProjectModel
                delegateSource: projectLine
                objectName: "projects"
                Layout.fillWidth: true

                property real maxProjectNameWidth: 0

                footer: Rectangle {
                    height: 50
                    width: table.width
                    border.width: 1
                    border.color: CSC.Style.lightGrey

                    RowLayout {
                        spacing: 30
                        anchors.fill: parent
                        anchors.leftMargin: CSC.Style.padding
                        anchors.rightMargin: CSC.Style.padding

                        Text {
                            id: levelText
                            text: "Name"
                            font.pixelSize: 13
                            font.weight: Font.Medium
                            Layout.fillWidth: true
                        }

                        Text {
                            text: "Location"
                            font.pixelSize: 13
                            font.weight: Font.Medium
                            visible: parent.width - table.maxProjectNameWidth > width + messageLabel.width + 2 * parent.spacing
                            Layout.preferredWidth: 150
                        }

                        Text {
                            id: messageLabel
                            text: "Progress"
                            font: textMetrics100.font
                            Layout.maximumWidth: 200
                            Layout.minimumWidth: 200
                        }
                    }
                }
            }

            Connections {
                target: QmlBridge
                onFuseReady: {
                    accessLayout.state = "finished"
                    tracker.progressIndex = 3
                    headerText.text = "<h1>Access ready</h1>"
                    infoText.text = "Data Gateway is ready to use"
                    buttonRow.visible = true
                }
            }
        }
    }  

    TextMetrics {
        id: textMetrics100
        text: "100 %"
        font.pixelSize: 13
        font.weight: Font.Medium
    }

    Component {
        id: projectLine

        Rectangle {
            height: 60
            width: table.width
            border.width: 1
            border.color: CSC.Style.lightGrey

            RowLayout {
                spacing: 30
                anchors.fill: parent
                anchors.leftMargin: CSC.Style.padding
                anchors.rightMargin: CSC.Style.padding

                Text {
                    text: projectName
                    font.pixelSize: 15
                    elide: Text.ElideRight
                    Layout.fillWidth: true

                    Component.onCompleted: {
                        if (implicitWidth > table.maxProjectNameWidth) {
                            table.maxProjectNameWidth = implicitWidth
                        }
                    }
                }

                Text {
                    text: repositoryName
                    font.pixelSize: 15
                    visible: parent.width - table.maxProjectNameWidth > width + loadingStatus.width + 2 * parent.spacing
                    Layout.preferredWidth: 150
                }

                RowLayout {
                    id: loadingStatus
                    Layout.maximumWidth: 200
                    Layout.minimumWidth: 200

                    property real value: (allContainers == -1) ? 0 : (allContainers == 0) ? 1 : loadedContainers / allContainers

                    CSC.ProgressBar {
                        id: progressbar
                        value: parent.value
                        Layout.fillWidth: true
                    }

                    Text {
                        id: percentValue
                        text: Math.floor(parent.value * 100) + "%"
                        //color: CSC.Style.grey
                        maximumLineCount: 1
                        font.pixelSize: 12
                        Layout.minimumWidth: textMetrics100.width
                    }
                }
            }
        }
    }  
}