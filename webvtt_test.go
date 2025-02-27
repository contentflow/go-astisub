package astisub_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/contentflow/go-astisub"
)

func TestWebVTT(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in.vtt")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Comments
	assert.Equal(t, []string{"this a nice example", "of a VTT"}, s.Items[0].Comments)
	assert.Equal(t, []string{"This a comment inside the VTT", "and this is the second line"}, s.Items[1].Comments)
	// Regions
	assert.Equal(t, 2, len(s.Regions))
	assert.Equal(t, astisub.Region{ID: "fred", InlineStyle: &astisub.StyleAttributes{WebVTTLines: 3, WebVTTRegionAnchor: "0%,100%", WebVTTScroll: "up", WebVTTViewportAnchor: "10%,90%", WebVTTWidth: "40%"}}, *s.Regions["fred"])
	assert.Equal(t, astisub.Region{ID: "bill", InlineStyle: &astisub.StyleAttributes{WebVTTLines: 3, WebVTTRegionAnchor: "100%,100%", WebVTTScroll: "up", WebVTTViewportAnchor: "90%,90%", WebVTTWidth: "40%"}}, *s.Regions["bill"])
	assert.Equal(t, s.Regions["bill"], s.Items[0].Region)
	assert.Equal(t, s.Regions["fred"], s.Items[1].Region)
	// Styles
	assert.Equal(t, astisub.StyleAttributes{WebVTTAlign: "left", WebVTTPosition: "10%,start", WebVTTSize: "35%"}, *s.Items[1].InlineStyle)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToWebVTT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.vtt")
	assert.NoError(t, err)
	err = s.WriteToWebVTT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}

func TestBroken1WebVTT(t *testing.T) {
	// Open bad, broken WebVTT file
	_, err := astisub.OpenFile("./testdata/broken-1-in.vtt")
	assert.Nil(t, err)
}

func TestWebVTTWithVoiceName(t *testing.T) {
	testData := `WEBVTT

	NOTE this a example with voicename

	1
	00:02:34.000 --> 00:02:35.000
	<v.first.local Roger Bingham> I'm the fist speaker

	2
	00:02:34.000 --> 00:02:35.000
	<v Bingham> I'm the second speaker

	3
	00:00:04.000 --> 00:00:08.000
	<v Lee>What are you doing here?</v>

	4
	00:00:04.000 --> 00:00:08.000
	<v Bob>Incorrect tag?</vi>`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	assert.NoError(t, err)

	assert.Len(t, s.Items, 4)
	assert.Equal(t, "Roger Bingham", s.Items[0].Lines[0].VoiceName)
	assert.Equal(t, "Bingham", s.Items[1].Lines[0].VoiceName)
	assert.Equal(t, "Lee", s.Items[2].Lines[0].VoiceName)
	assert.Equal(t, "Bob", s.Items[3].Lines[0].VoiceName)

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	assert.NoError(t, err)
	assert.Equal(t, `WEBVTT

NOTE this a example with voicename

1
00:02:34.000 --> 00:02:35.000
<v Roger Bingham>I'm the fist speaker

2
00:02:34.000 --> 00:02:35.000
<v Bingham>I'm the second speaker

3
00:00:04.000 --> 00:00:08.000
<v Lee>What are you doing here?

4
00:00:04.000 --> 00:00:08.000
<v Bob>Incorrect tag?
`, b.String())
}

func TestWebVTTWithTimestampMap(t *testing.T) {
	testData := `WEBVTT
	X-TIMESTAMP-MAP=MPEGTS:180000, LOCAL:00:00:00.000

	00:00.933 --> 00:02.366
	♪ ♪

	00:02.400 --> 00:03.633
	Evening.`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	assert.NoError(t, err)

	assert.Len(t, s.Items, 2)

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	assert.NoError(t, err)
	assert.Equal(t, `WEBVTT

1
00:00:02.933 --> 00:00:04.366
♪ ♪

2
00:00:04.400 --> 00:00:05.633
Evening.
`, b.String())
}

func TestWebVTTEscape(t *testing.T) {
	testData := `WEBVTT

	00:01:00.000 --> 00:02:00.000
	Sentence with an &amp; in the middle

	00:02:00.000 --> 00:03:00.000
	Sentence with an &lt; in the middle`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	require.NoError(t, err)

	require.Len(t, s.Items, 2)
	require.Equal(t, "Sentence with an & in the middle", s.Items[0].String())
	require.Equal(t, "Sentence with an < in the middle", s.Items[1].String())

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	require.NoError(t, err)
	require.Equal(t, `WEBVTT

1
00:01:00.000 --> 00:02:00.000
Sentence with an &amp; in the middle

2
00:02:00.000 --> 00:03:00.000
Sentence with an &lt; in the middle
`, b.String())
}

func TestWebVTTTags(t *testing.T) {
	testData := `WEBVTT

	00:01:00.000 --> 00:02:00.000
	<u><i>Italic with underline text</i></u> some extra

	00:02:00.000 --> 00:03:00.000
	<lang en>English here</lang> <c.yellow.bg_blue>Yellow text on blue background</c>

	00:03:00.000 --> 00:04:00.000
	<v Joe><c.red><i>Joe's words are red in italic</i></c>

	00:04:00.000 --> 00:05:00.000
	<customed_tag.class1.class2>Text here</customed_tag>

	00:05:00.000 --> 00:06:00.000
	<v Joe>Joe says something</v> <v Bob>Bob says something</v>`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	require.NoError(t, err)

	require.Len(t, s.Items, 5)

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	require.NoError(t, err)
	require.Equal(t, `WEBVTT

1
00:01:00.000 --> 00:02:00.000
<u><i>Italic with underline text</i></u> some extra

2
00:02:00.000 --> 00:03:00.000
<lang en>English here</lang> <c.yellow.bg_blue>Yellow text on blue background</c>

3
00:03:00.000 --> 00:04:00.000
<v Joe><c.red><i>Joe's words are red in italic</i></c>

4
00:04:00.000 --> 00:05:00.000
<customed_tag.class1.class2>Text here</customed_tag>

5
00:05:00.000 --> 00:06:00.000
<v Joe>Joe says something Bob says something
`, b.String())
}
