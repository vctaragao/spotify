 /*
  mpg123_to_wav.c
 
  This is example code only sensible to be considered in the public domain.
  Initially written by Nicholas Humfrey.
 
  The most complicated part is about the choices to make about output format,
  and prepare for the unlikely case a bastard mp3 might file change it.
 */
 
 #include <out123.h>
 #include <mpg123.h>
 #include <stdio.h>
 #include <strings.h>


#define SECONDS 16  

 void usage(const char *cmd)
 {
  printf("Usage: %s <input> [<offset> [<driver> [<output> [encoding [buffersize]]]]]\n" , cmd);
  printf( "\nPlay MPEG audio from intput file to output file/device using\n"
  "specified out123 driver, sample encoding and buffer size optional.\n\n" );
  exit(99);
 }
 
 void cleanup(mpg123_handle *mh, out123_handle *ao)
 {
  out123_del(ao);
  /* It's really to late for error checks here;-) */
  mpg123_close(mh);
  mpg123_delete(mh);
 }
 
 int main(int argc, char *argv[])
 {
  mpg123_handle *mh = NULL;
  out123_handle *ao = NULL;
  char *infile = NULL;
  char *driver = NULL;
  char *outfile = NULL;
  unsigned char* buffer = NULL;
  const char *encname;
  size_t buffer_size = 0;
  size_t buffer_size_default = 0;
  size_t done = 0;
  int channels = 0;
  int encoding = 0;
  int framesize = 1;
  long rate = 0;
  int err = MPG123_OK;
  off_t samples = 0;

  off_t offset = 0;
  int offset_seconds = 0;
 
  if(argc<2) usage(argv[0]);
 
  infile = argv[1];

  if(argc >= 3) offset_seconds = atoi(argv[2]);
  printf("[debug] offset_seconds: %d\n" , offset_seconds);

  if(argc >= 4) outfile = argv[3];

  printf("Input file: %s\n", infile);
  printf("Output driver: %s\n", driver ? driver : "<nil> (default)");
  printf("Output file: %s\n", outfile ? outfile : "<nil> (default)");

#if MPG123_API_VERSION < 46
  // Newer versions of the library don't need that anymore, but it is safe
  // to have the no-op call present for compatibility with old versions.
  err = mpg123_init();
#endif

  mh = mpg123_new(NULL, &err);
  if(err != MPG123_OK || mh == NULL)
  {
      fprintf(stderr, "Basic setup goes wrong: %s", mpg123_plain_strerror(err));
      cleanup(mh, ao);
      return -1;
  }

  ao = out123_new();
  if(!ao)
  {
      fprintf(stderr, "Cannot create output handle.\n");
      cleanup(mh, ao);
      return -1;
  }

  if(argc >= 5)
  { /* Make mpg123 support the desired encoding only for all rates. */
      const long *rates;
      size_t rate_count;
      size_t i;
      int enc;
      /* If that is zero, you'll get the error soon enough from mpg123. */
      enc = out123_enc_byname(argv[4]);
      mpg123_format_none(mh);
      mpg123_rates(&rates, &rate_count);
      for(i=0; i<rate_count; ++i)
          mpg123_format(mh, rates[i], MPG123_MONO|MPG123_STEREO, enc);
  }

  /* Let mpg123 work with the file, that excludes MPG123_NEED_MORE messages. */
  if( mpg123_open(mh, infile) != MPG123_OK || mpg123_getformat(mh, &rate, &channels, &encoding) != MPG123_OK )
  {
      fprintf( stderr, "Trouble with mpg123: %s\n", mpg123_strerror(mh) );
      cleanup(mh, ao);
      return -1;
  }

  if(out123_open(ao, driver, outfile) != OUT123_OK)
  {
      fprintf(stderr, "Trouble with out123: %s\n", out123_strerror(ao));
      cleanup(mh, ao);
      return -1;
  }

  /* It makes no sense for that to give an error now. */
  out123_driver_info(ao, &driver, &outfile);
  printf("Effective output driver: %s\n", driver ? driver : "<nil> (default)");
  printf("Effective output file: %s\n", outfile ? outfile : "<nil> (default)");

  /* Ensure that this output format will not change (it might, when we allow it). */
  mpg123_format_none(mh);
  mpg123_format(mh, rate, channels, encoding);

  encname = out123_enc_name(encoding);
  printf( "Playing with %i channels and %li Hz, encoding %s.\n" , channels, rate, encname ? encname : "???" );
  if( out123_start(ao, rate, channels, encoding) || out123_getformat(ao, NULL, NULL, NULL, &framesize) )
  {
      fprintf(stderr, "Cannot start output / get framesize: %s\n" , out123_strerror(ao));
      cleanup(mh, ao);
      return -1;
  }

  long one_second = rate * framesize;
  printf("[debug] one_second: %ld\n", one_second); // 176_400
                                                   
  /* seek from desired offset */
  off_t sample_off = one_second * offset_seconds;
  printf("[debug] sample_off: %ld\n", sample_off); // 2_822_400
  offset = mpg123_seek(mh, sample_off, SEEK_SET);

  if(offset < 0 ){
      fprintf(stderr, "Offset not able to be reached: %s\n" , mpg123_strerror(mh));
      cleanup(mh, ao);
      return -1;
  }
  printf("[debug] offset: %ld\n", offset); // 2_822_400
                                           
  /* Buffer could be almost any size here, mpg123_outblock() is just some
   * recommendation. The size should be a multiple of the PCM frame size.
   * 1s of music in buffer = default_buffer * ( 1s of music / default_buffer) 
   * 15s of music in buffer = 1s of music in buffer * 15
   */
  buffer_size_default = argc >= 6 ? atol(argv[5]) : mpg123_outblock(mh);
  printf("[debug] buffer_size_default: %ld\n", buffer_size_default); // 4_608 
  buffer_size = buffer_size_default * ((one_second / buffer_size_default) * SECONDS);

  printf("[debug] buffer_size: %ld\n", buffer_size); // 2_646_000
  buffer = malloc( buffer_size );

  //  do
  //  {
  //      size_t played;
  //      err = mpg123_read( mh, buffer, buffer_size, &done );
  //      played = out123_play(ao, buffer, done);
  //      if(played != done)
  //      {
  //          fprintf(stderr
  //                  , "Warning: written less than gotten from libmpg123: %li != %li\n"
  //                  , (long)played, (long)done);
  //      }
  //      printf("[debug] played: %ld\n", played);
  //      printf("[debug] framesize: %d\n", framesize);
  //      samples += played/framesize;
  //      printf("[debug] sample: %ld\n\n", samples);
  //      /* We are not in feeder mode, so MPG123_OK, MPG123_ERR and
  //         MPG123_NEW_FORMAT are the only possibilities.
  //         We do not handle a new format, MPG123_DONE is the end... so
  //         abort on anything not MPG123_OK. */
  //  } while (done && err==MPG123_OK);

  size_t played;
  err = mpg123_read( mh, buffer, buffer_size, &done );
  printf("[debug] done: %ld\n", done);
  played = out123_play(ao, buffer, done);
  if(played != done)
  {
      fprintf(stderr
              , "Warning: written less than gotten from libmpg123: %li != %li\n"
              , (long)played, (long)done);
  }
  printf("[debug] played: %ld\n", played);
  printf("[debug] framesize: %d\n", framesize);
  samples += played/framesize;
  printf("[debug] sample: %ld\n\n", samples);

  free(buffer);

  if(err != MPG123_DONE)
      fprintf( stderr, "Warning: Decoding ended prematurely because: %s\n", err == MPG123_ERR ? mpg123_strerror(mh) : mpg123_plain_strerror(err) );

  printf("%li samples written.\n", (long)samples);
  cleanup(mh, ao);
  return 0;
 }
