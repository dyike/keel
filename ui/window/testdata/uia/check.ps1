# Reads a running Keel app through UI Automation's managed client, from
# another process, as Narrator does.
#   check.ps1 hello <exe>       toggles the checkbox, invokes "Say hello"
#   check.ps1 components <exe>  scrolls the "Tree" link into view, invokes it
param([string]$Mode, [string]$Exe)
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes
$AE = [System.Windows.Automation.AutomationElement]
$Scope = [System.Windows.Automation.TreeScope]
$CT = [System.Windows.Automation.ControlType]

function Wait-For([scriptblock]$Probe, [string]$What) {
  for ($i = 0; $i -lt 150; $i++) {
    $r = & $Probe
    if ($r) { return $r }
    Start-Sleep -Milliseconds 100
  }
  throw "FAIL: timed out waiting for $What"
}
function Find([object]$Under, [object]$Type, [string]$Name) {
  $c = New-Object System.Windows.Automation.AndCondition(
    (New-Object System.Windows.Automation.PropertyCondition($AE::ControlTypeProperty, $Type)),
    (New-Object System.Windows.Automation.PropertyCondition($AE::NameProperty, $Name)))
  return $Under.FindFirst($Scope::Descendants, $c)
}
function Dump([object]$Win) {
  foreach ($e in $Win.FindAll($Scope::Descendants, [System.Windows.Automation.Condition]::TrueCondition)) {
    "{0} '{1}' offscreen={2}" -f $e.Current.ControlType.ProgrammaticName, $e.Current.Name, $e.Current.IsOffscreen
  }
}

$proc = Start-Process $Exe -PassThru
try {
  $byPid = New-Object System.Windows.Automation.PropertyCondition($AE::ProcessIdProperty, $proc.Id)
  $win = Wait-For { $AE::RootElement.FindFirst($Scope::Children, $byPid) } 'the window'
  if ($Mode -eq 'hello') {
    $box = Wait-For { Find $win $CT::CheckBox 'Add a welcome message' } 'the checkbox'
    Dump $win
    $toggle = $box.GetCurrentPattern([System.Windows.Automation.TogglePattern]::Pattern)
    if ($toggle.Current.ToggleState -ne 'Off') { throw 'FAIL: checkbox starts on' }
    $toggle.Toggle()
    Wait-For { $toggle.Current.ToggleState -eq 'On' } 'the checkbox to turn on' | Out-Null
    $button = Find $win $CT::Button 'Say hello'
    $button.GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern).Invoke()
    Wait-For { Find $win $CT::Text 'Enter your name first.' } 'the status text' | Out-Null
    'PASS: UI Automation reads the tree, toggles the checkbox and invokes the button'
  } else {
    $link = Wait-For { Find $win $CT::Hyperlink 'Tree' } 'the Tree link'
    if (-not $link.Current.IsOffscreen) { throw 'FAIL: Tree link shows before scrolling' }
    $link.GetCurrentPattern([System.Windows.Automation.ScrollItemPattern]::Pattern).ScrollIntoView()
    Wait-For { -not (Find $win $CT::Hyperlink 'Tree').Current.IsOffscreen } 'the Tree link to scroll into view' | Out-Null
    (Find $win $CT::Hyperlink 'Tree').GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern).Invoke()
    # The page's subtitle names its doc (ASCII only: Windows PowerShell 5.1
    # reads this file as ANSI).
    Wait-For { $win.FindAll($Scope::Descendants, [System.Windows.Automation.Condition]::TrueCondition) |
        Where-Object { $_.Current.Name -like '*docs/kit/tree.md' } } 'the Tree page' | Out-Null
    'PASS: UI Automation scrolls a link out of view into view and invokes it'
  }
} catch {
  if ($win) { Dump $win }
  throw
} finally {
  Stop-Process -Id $proc.Id -ErrorAction SilentlyContinue
}
