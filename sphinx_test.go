package gosphinx

import (
	"fmt"
	"testing"
)

var (
	sc *SphinxClient
	//host = "/var/run/searchd.sock"
	host = "0.0.0.0"
	port = 9312 // If set host to unix path, then just ignore port.
	index = "test1"
	words = "test"
)


func TestMultiQuery(t *testing.T) {
	fmt.Println("Running multi Query() test...")
	f := func(i int){
		scMulti := NewSphinxClient()
	    scMulti.SetServer(host, port)
	    if err := scMulti.Open(); err != nil {
	    	t.Fatalf("Multi Query test, Open() > %v\n", err)
	    }
	    defer scMulti.Close()
		
		res, err := scMulti.Query(words, index, "test multi Query()")
		if err != nil {
			t.Fatalf("Multi Query %d -> %s\n", i, err)
		}
		
		if res.Total != 3 || res.TotalFound != 3 {
			t.Fatalf("Mulit Query %d -> res.Total: %d\tres.TotalFound: %d\n", i, res.Total, res.TotalFound)
		}
		
		if scMulti.GetLastWarning() != "" {
			fmt.Printf("Mulit Query %d warning: %s\n", i, scMulti.GetLastWarning())
		}
	}
	
	for i:=0;i<100;i++ {
		if i > 0 && i%10 == 0 {
			fmt.Printf("Already start %d goroutines...\n", i)
		}
		go f(i)
	}
}

func TestInitClient(t *testing.T) {
	fmt.Println("Init sphinx client ...")
    sc = NewSphinxClient()
    sc.SetServer(host, port)
    if err := sc.Open(); err != nil {
    	t.Fatalf("Init sphinx client > %v\n", err)
    }
    
    status, err := sc.Status()
	if err != nil {
		t.Fatalf("Error: %s\n", err)
		return
	}
	
	for _, row := range status {
		fmt.Printf("%20s:\t%s\n", row[0], row[1])
	}
}


func TestQuery(t *testing.T) {
	fmt.Println("Running sphinx Query() test...")
    
	res, err := sc.Query(words, index, "test Query()")
	if err != nil {
		t.Fatalf("Query -> %s\n", err)
	}
	
	if res.Total != 3 || res.TotalFound != 3 {
		t.Fatalf("Query -> res.Total: %d\tres.TotalFound: %d\n", res.Total, res.TotalFound)
	}
	
	if sc.GetLastWarning() != "" {
		fmt.Printf("Query warning: %s\n", sc.GetLastWarning())
	}
}

func TestAddQueryAndRunQueries(t *testing.T){
	fmt.Println("Running sphinx AddQuery() and RunQueries() test...")
	_, err := sc.AddQuery("my", index, "It's the second Query.")
	
	results, err := sc.RunQueries()
	if err != nil {
		t.Fatalf("RunQueries -> %s\n", err)
	}
	
	// TestQuery already add one.
	if len(results) != 2 {
		t.Fatalf("RunQueries -> get %d results.\n", len(results))
		
		for i, res := range results {
			fmt.Printf("%dth result: %#v\n", i, res)
		}
	}
}

// If you do not use xml data source, just comment this func.
func TestQueryXml(t *testing.T) {
	fmt.Println("Running sphinx Query() xml test...")
	
	// Test word "understand" in index "xml"
	res, err := sc.Query("understand", "xml", "test xml Query()")
	if err != nil {
		t.Fatalf("Query xml -> %s\n", err)
	}
	
	if res.Total != 1 || res.TotalFound != 1 {
		t.Fatalf("Query xml -> res.Total: %d\tres.TotalFound: %d\n", res.Total, res.TotalFound)
	}
	
	if res.Matches[0].DocId != 1235 {
		t.Fatalf("Query xml -> res.Matches: %v\n", res.Matches)
	}

	if sc.GetLastWarning() != "" {
		fmt.Printf("Query xml warning: %s\n", sc.GetLastWarning())
	}
}

func TestBuildExcerpts(t *testing.T) {
	fmt.Println("Running sphinx BuildExcerpts() test...")
	docs := []string{
		"this is my first test text to be highlighted, and for the sake of the testing we need to pump its length somewhat",
		"another test text to be highlighted, below limit",
		"number three, without phrase match",
		"final test, not only without phrase match, but also above limit and with swapped phrase text test as well",
	}

	opts := ExcerptsOpts{
		BeforeMatch : "<span style='font-weight:bold;color:red'>",
		AfterMatch : "</span>",
		ChunkSeparator : " ... ",
		Limit : 60,
		Around : 3,
	}
	
	res, err := sc.BuildExcerpts(docs, index, words, opts)
	if err != nil {
		t.Fatalf("BuildExcerpts -> %s\n", err)
	}
	
	if res[0] != ` ...  is my first <span style='font-weight:bold;color:red'>test</span> text to be ... ` {
		t.Fatalf("BuildExcerpts res[0]: %#v\n", res[0])
	}
	if res[1] != `another <span style='font-weight:bold;color:red'>test</span> text to be highlighted, below limit` {
		t.Fatalf("BuildExcerpts res[1]: %#v\n", res[1])
	}
	if res[2] != `number three, without phrase match` {
		t.Fatalf("BuildExcerpts res[2]: %#v\n", res[2])
	}
	if res[3] != `final <span style='font-weight:bold;color:red'>test</span>, not only without  ...  swapped phrase text <span style='font-weight:bold;color:red'>test</span> as well` {
		t.Fatalf("BuildExcerpts res[3]: %#v\n", res[3])
	}
}

func TestUpdateAttributes(t *testing.T){
	fmt.Println("Running sphinx UpdateAttributes() test...")
	//UpdateAttributes(index string, attrs []string, values [][]interface{}) (ndocs int, err error)
	sc = NewSphinxClient()
	sc.SetServer(host, port)
	attrs := []string{"group_id", "group_id2"} //, "cate_ids"
	v1 := []interface{}{uint64(1), 3, 15}
	v2 := []interface{}{uint64(2), 4, 16}
	values := [][]interface{}{v1, v2}
	//v3 := []interface{uint64(4), []int{4,5,6,7}}
	ndocs, err := sc.UpdateAttributes(index, attrs, values)
	if err != nil {
		t.Fatalf("UpdateAttributes -> %s\n", err)
	}
	
	if ndocs != 2 {
		t.Fatalf("UpdateAttributes -> ndocs: %d\n", ndocs)
	}
	
	sc.SetFilter("group_id", []uint64{3,4}, true) // exclude 3,4, then should get doc3 and doc4.
	result, err := sc.Query("", index, "")
	if err != nil {
		t.Fatalf("UpdateAttributes -> Query -> %#v\n", err)
	}
	
	if result.Total != 2 {
		t.Fatalf("UpdateAttributes -> total: %d\n", result.Total)
	}
	
	if result.Matches[0].DocId != 3 || result.Matches[1].DocId != 4 {
		t.Fatalf("UpdateAttributes -> wrong matches: %#v\n", result.Matches)
	}
}

func TestBuildKeywords(t *testing.T) {
	fmt.Println("Running sphinx BuildKeywords() test...")
	keywords, err := sc.BuildKeywords("this.is.my query", index, true)
	if err != nil {
		t.Fatalf("BuildKeywords -> %s\n", err)
	}
	
	if len(keywords) != 4 {
		t.Fatalf("BuildKeywords -> just get %d keywords! actually 4 keywords.\n", len(keywords))
		
		for i, kw := range keywords {
			fmt.Printf("Keywords %d : %#v\n", i, kw)
		}
	}
}

