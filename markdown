=== GO SOURCE CODE STATISTICS REPORT ===
Repository: /home/user/go/src/github.com/opd-ai/go-stats-generator/testdata/simple
Generated: 2025-07-25 20:41:53
Analysis Time: 2ms
Files Processed: 6

=== OVERVIEW ===
Total Lines of Code: 0
Total Functions: 14
Total Methods: 15
Total Structs: 4
Total Interfaces: 5
Total Packages: 3
Total Files: 6

=== FUNCTION ANALYSIS ===
Function Statistics:
  Average Function Length: 19.6 lines
  Longest Function: VeryComplexFunction (84 lines)
  Functions > 50 lines: 3 (10.3%)
  Functions > 100 lines: 0 (0.0%)
  Average Complexity: 7.9
  High Complexity (>10): 2 functions

Top Complex Functions:
Rank Function                  File                    Lines Complexity
-----------------------------------------------------------------------
   1 VeryComplexFunction       simple                     84       46.2
   2 ComplexFunction           simple                     43       27.0
   3 FanOutExample             concurrency                51       19.0
   4 FanInExample              concurrency                62       17.9
   5 ContextCancellationExa... concurrency                29       15.2
   6 WorkerPoolExample         concurrency                34       10.5
   7 PotentialLeakExample      concurrency                19        9.5
   8 Statistics                simple                     43        9.5
   9 PipelineExample           concurrency                22        8.2
  10 UpdateUser                simple                     14        7.2

=== COMPLEXITY ANALYSIS ===
Top 29 Most Complex Functions:
Function                       Package                 Lines Cyclomatic    Overall
--------------------------------------------------------------------------------
VeryComplexFunction            simple                     84         24       46.2
ComplexFunction                simple                     43         15       27.0
FanOutExample                  concurrency                51         10       19.0
FanInExample                   concurrency                62          8       17.9
ContextCancellationExample     concurrency                29          9       15.2
WorkerPoolExample              concurrency                34          5       10.5
PotentialLeakExample           concurrency                19          5        9.5
Statistics                     simple                     43          5        9.5
PipelineExample                concurrency                22          4        8.2
UpdateUser                     simple                     14          4        7.2
CreateUser                     simple                     18          3        5.4
GetActiveUsers                 simple                      7          3        5.4
SemaphoreExample               concurrency                18          2        4.6
ComplexLineCountingTest        test                       17          2        3.6
Divide                         simple                      7          2        3.6
GetUser                        simple                      5          2        3.6
GetAllUsers                    simple                      5          2        3.6
DeleteUser                     simple                      7          2        3.6
DeactivateUser                 simple                      7          2        3.6
SyncPrimitivesExample          concurrency                39          1        3.3
SimpleGoroutineTest            simple                     13          1        2.8
Add                            simple                      3          1        1.8
Subtract                       simple                      3          1        1.8
Multiply                       simple                      3          1        1.8
NewCalculator                  simple                      3          1        1.8
recordOperation                simple                      6          1        1.8
GetHistory                     simple                      1          1        1.8
ClearHistory                   simple                      1          1        1.8
NewUserService                 simple                      3          1        1.8

=== PACKAGE ANALYSIS ===
Total Packages: 3
Average Dependencies per Package: 0.0
Average Files per Package: 2.0

Low Cohesion Packages (<2.0 cohesion score):
  concurrency: 1.6 cohesion, 1 files, 8 functions
  simple: 1.4 cohesion, 4 files, 20 functions
  test: 0.2 cohesion, 1 files, 1 functions

Largest Packages (by function count):
  simple: 20 functions, 9 structs, 0 interfaces, 4 files
  concurrency: 8 functions, 0 structs, 0 interfaces, 1 files
  test: 1 functions, 0 structs, 0 interfaces, 1 files

=== ANALYSIS COMPLETE ===
Report generated by go-stats-generator v1.0.0
