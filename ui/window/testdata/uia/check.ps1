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

# Wait-For retries a probe; one that walks the tree while the window
# replaces elements may meet one that just left (ElementNotAvailable).
function Wait-For([scriptblock]$Probe, [string]$What) {
  for ($i = 0; $i -lt 150; $i++) {
    try {
      $r = & $Probe
      if ($r) { return $r }
    } catch [System.Windows.Automation.ElementNotAvailableException] {
    }
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
# Windows of the process: it may own more than one top-level window, and
# the one to check is the one holding the element sought.
function Windows-Of([int]$ProcessId) {
  $byPid = New-Object System.Windows.Automation.PropertyCondition($AE::ProcessIdProperty, $ProcessId)
  return @($AE::RootElement.FindAll($Scope::Children, $byPid))
}
function Find-Any([object]$Type, [string]$Name) {
  foreach ($w in Windows-Of $proc.Id) {
    $e = Find $w $Type $Name
    if ($e) { $script:win = $w; return $e }
  }
  return $null
}
function Dump-All() {
  foreach ($w in Windows-Of $proc.Id) {
    $all = $w.FindAll($Scope::Descendants, [System.Windows.Automation.Condition]::TrueCondition)
    "window '{0}' class '{1}': {2} elements" -f $w.Current.Name, $w.Current.ClassName, $all.Count
  }
}
function Dump([object]$Win) {
  foreach ($e in $Win.FindAll($Scope::Descendants, [System.Windows.Automation.Condition]::TrueCondition)) {
    "{0} '{1}' offscreen={2}" -f $e.Current.ControlType.ProgrammaticName, $e.Current.Name, $e.Current.IsOffscreen
  }
}

$proc = Start-Process $Exe -PassThru
try {
  $win = $null
  if ($Mode -eq 'hello') {
    $box = Wait-For { Find-Any $CT::CheckBox 'Add a welcome message' } 'the checkbox'
    Dump $win
    $toggle = $box.GetCurrentPattern([System.Windows.Automation.TogglePattern]::Pattern)
    if ($toggle.Current.ToggleState -ne 'Off') { throw 'FAIL: checkbox starts on' }
    $toggle.Toggle()
    Wait-For { $toggle.Current.ToggleState -eq 'On' } 'the checkbox to turn on' | Out-Null
    $button = Find $win $CT::Button 'Say hello'
    $button.GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern).Invoke()
    Wait-For { Find $win $CT::Text 'Enter your name first.' } 'the status text' | Out-Null
    # The press also gave the button keyboard focus.
    Wait-For { (Find $win $CT::Button 'Say hello').Current.HasKeyboardFocus } 'the button to hold focus' | Out-Null
    # A client moves focus with SetFocus.
    $clear = Find $win $CT::Button 'Clear'
    $clear.SetFocus()
    Wait-For { (Find $win $CT::Button 'Clear').Current.HasKeyboardFocus } 'SetFocus to move focus to Clear' | Out-Null
    'PASS: UI Automation reads the tree, toggles the checkbox, invokes the button, and reports and sets focus'
  } else {
    $link = Wait-For { Find-Any $CT::Hyperlink 'Tree' } 'the Tree link'
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
  try { Dump-All; if ($win) { Dump $win } } catch {}
  throw
} finally {
  Stop-Process -Id $proc.Id -ErrorAction SilentlyContinue
}
